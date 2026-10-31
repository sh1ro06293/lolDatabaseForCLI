package projectlint

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	messageTypeDecl     = "型の宣言は model.go か *_config.go に置く"
	messageErrVar       = "エラーの値は errors.go に置く"
	messageComment      = "コメントを「。」で終えない"
	messageErrorString  = "エラー文を「。」で終えない"
	messageTestRequired = "internal と cmd のパッケージには _test.go を置く"

	sentenceEnd = "。"
)

func Check(fsys fs.FS) ([]Finding, error) {
	names, err := goFiles(fsys)
	if err != nil {
		return nil, fmt.Errorf("ソースを読めません: %w", err)
	}

	var findings []Finding

	for _, name := range names {
		fileFindings, checkErr := checkFile(fsys, name)
		if checkErr != nil {
			return nil, checkErr
		}

		findings = append(findings, fileFindings...)
	}

	findings = append(findings, checkTestFiles(names)...)

	slices.SortFunc(findings, compareFindings)

	return findings, nil
}

func goFiles(fsys fs.FS) ([]string, error) {
	var names []string

	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if skipDir(name) {
				return fs.SkipDir
			}

			return nil
		}

		// シンボリックリンクを読むと、リポジトリの外のファイルを検査したり、リンク切れで検査全体が止まったりする
		if entry.Type().IsRegular() && path.Ext(name) == ".go" {
			names = append(names, name)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("ファイルを探せません: %w", err)
	}

	return names, nil
}

func checkTestFiles(names []string) []Finding {
	firstFiles := make(map[string]string)
	hasTest := make(map[string]bool)

	var dirs []string

	for _, name := range names {
		dir := path.Dir(name)
		if !strings.HasPrefix(dir, "internal/") && !strings.HasPrefix(dir, "cmd/") {
			continue
		}

		if strings.HasSuffix(name, "_test.go") {
			hasTest[dir] = true

			continue
		}

		if _, seen := firstFiles[dir]; !seen {
			firstFiles[dir] = name
			dirs = append(dirs, dir)
		}
	}

	var findings []Finding

	for _, dir := range dirs {
		if !hasTest[dir] {
			findings = append(findings, Finding{
				Path:    firstFiles[dir],
				Line:    1,
				Column:  1,
				Message: messageTestRequired,
			})
		}
	}

	return findings
}

func skipDir(name string) bool {
	base := path.Base(name)
	if name == "." {
		return false
	}

	return strings.HasPrefix(base, ".") || base == "testdata" || base == "vendor"
}

func checkFile(fsys fs.FS, name string) ([]Finding, error) {
	src, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("%s を読めません: %w", name, err)
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, name, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("%s を解析できません: %w", name, err)
	}

	var findings []Finding

	base := path.Base(name)
	if !strings.HasSuffix(base, "_test.go") {
		findings = append(findings, checkTypeDecls(fset, file, base)...)
		findings = append(findings, checkErrVars(fset, file, base)...)
	}

	findings = append(findings, checkComments(fset, file)...)
	findings = append(findings, checkErrorStrings(fset, file)...)

	return findings, nil
}

func checkTypeDecls(fset *token.FileSet, file *ast.File, base string) []Finding {
	if base == "model.go" || strings.HasSuffix(base, "_config.go") {
		return nil
	}

	var findings []Finding

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			findings = append(findings, newFinding(fset, spec.Pos(), messageTypeDecl))
		}
	}

	return findings
}

func checkErrVars(fset *token.FileSet, file *ast.File, base string) []Finding {
	if base == "errors.go" {
		return nil
	}

	var findings []Finding

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, ident := range value.Names {
				if isErrVarName(ident.Name) {
					findings = append(findings, newFinding(fset, ident.Pos(), messageErrVar))
				}
			}
		}
	}

	return findings
}

// Errata のように Err で始まるだけの名前を除くため、Err の次が大文字かを見る
func isErrVarName(name string) bool {
	rest, found := strings.CutPrefix(name, "Err")
	if !found {
		return false
	}

	if rest == "" {
		return true
	}

	first, _ := utf8.DecodeRuneInString(rest)

	return unicode.IsUpper(first)
}

func checkComments(fset *token.FileSet, file *ast.File) []Finding {
	var findings []Finding

	for _, group := range file.Comments {
		for _, comment := range group.List {
			start := fset.PositionFor(comment.Slash, false)
			text := strings.TrimSuffix(comment.Text, "*/")

			for i, line := range strings.Split(text, "\n") {
				if !strings.HasSuffix(strings.TrimSpace(line), sentenceEnd) {
					continue
				}

				pos := start
				pos.Line += i

				if i > 0 {
					pos.Column = 1
				}

				findings = append(findings, Finding{
					Path:    pos.Filename,
					Line:    pos.Line,
					Column:  pos.Column,
					Message: messageComment,
				})
			}
		}
	}

	return findings
}

func checkErrorStrings(fset *token.FileSet, file *ast.File) []Finding {
	var findings []Finding

	constructors := errorConstructors(file)

	ast.Inspect(file, func(node ast.Node) bool {
		call, isCall := node.(*ast.CallExpr)
		if !isCall || !isErrorConstructor(call, constructors) || len(call.Args) == 0 {
			return true
		}

		lit, isLit := call.Args[0].(*ast.BasicLit)
		if !isLit || lit.Kind != token.STRING {
			return true
		}

		text, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		if strings.HasSuffix(strings.TrimSpace(text), sentenceEnd) {
			findings = append(findings, newFinding(fset, lit.Pos(), messageErrorString))
		}

		return true
	})

	return findings
}

// errors と fmt は別名で読み込まれることがあるので、名前ではなく import から関数を決める
func errorConstructors(file *ast.File) map[string]string {
	constructors := make(map[string]string)

	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		funcName := errorConstructorFunc(importPath)
		if funcName == "" {
			continue
		}

		localName := importPath
		if spec.Name != nil {
			localName = spec.Name.Name
		}

		constructors[localName] = funcName
	}

	return constructors
}

func errorConstructorFunc(importPath string) string {
	switch importPath {
	case "errors":
		return "New"
	case "fmt":
		return "Errorf"
	default:
		return ""
	}
}

func isErrorConstructor(call *ast.CallExpr, constructors map[string]string) bool {
	sel, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	pkg, isIdent := sel.X.(*ast.Ident)
	if !isIdent {
		return false
	}

	funcName, found := constructors[pkg.Name]

	return found && sel.Sel.Name == funcName
}

func newFinding(fset *token.FileSet, pos token.Pos, message string) Finding {
	// //line の指示に従うと、実際のファイルではなく指示先の位置を報告してしまう
	position := fset.PositionFor(pos, false)

	return Finding{
		Path:    position.Filename,
		Line:    position.Line,
		Column:  position.Column,
		Message: message,
	}
}

func compareFindings(a, b Finding) int {
	return cmp.Or(
		cmp.Compare(a.Path, b.Path),
		cmp.Compare(a.Line, b.Line),
		cmp.Compare(a.Column, b.Column),
	)
}
