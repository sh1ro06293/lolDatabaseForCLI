package projectlint_test

import (
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/projectlint"
)

type testCase struct {
	name  string
	files fstest.MapFS
	want  []projectlint.Finding
}

func file(src string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(src)}
}

func runCases(t *testing.T, tests []testCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := projectlint.Check(tt.files)
			if err != nil {
				t.Fatalf("Check() のエラー: %v", err)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("Check() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCheckTypeDecls(t *testing.T) {
	t.Parallel()

	runCases(t, []testCase{
		{
			name: "model.go と *_config.go とテストには置ける",
			files: fstest.MapFS{
				"a/model.go":         file("package a\n\ntype A struct{}\n"),
				"a/client_config.go": file("package a\n\ntype Client struct{}\n"),
				"a/check_test.go":    file("package a\n\ntype helper struct{}\n"),
			},
			want: nil,
		},
		{
			name: "他のファイルに置くと違反",
			files: fstest.MapFS{
				"a/check.go": file("package a\n\ntype (\n\tA struct{}\n\tB int\n)\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/check.go", Line: 4, Column: 2, Message: "型の宣言は model.go か *_config.go に置く"},
				{Path: "a/check.go", Line: 5, Column: 2, Message: "型の宣言は model.go か *_config.go に置く"},
			},
		},
		{
			name: "関数の中の型は対象外",
			files: fstest.MapFS{
				"a/check.go": file("package a\n\nfunc f() {\n\ttype local struct{}\n\n\t_ = local{}\n}\n"),
			},
			want: nil,
		},
	})
}

func TestCheckErrVars(t *testing.T) {
	t.Parallel()

	runCases(t, []testCase{
		{
			name: "errors.go とテストには置ける",
			files: fstest.MapFS{
				"a/errors.go":     file("package a\n\nimport \"errors\"\n\nvar ErrA = errors.New(\"a\")\n"),
				"a/check_test.go": file("package a\n\nimport \"errors\"\n\nvar ErrTest = errors.New(\"t\")\n"),
			},
			want: nil,
		},
		{
			name: "他のファイルに置くと違反",
			files: fstest.MapFS{
				"a/check.go": file("package a\n\nimport \"errors\"\n\nvar ErrA = errors.New(\"a\")\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/check.go", Line: 5, Column: 5, Message: "エラーの値は errors.go に置く"},
			},
		},
		{
			name: "Err の次が大文字でなければ対象外",
			files: fstest.MapFS{
				"a/check.go": file("package a\n\nvar (\n\tErrata = 1\n\tErrorMessages = 2\n\tErr = 3\n)\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/check.go", Line: 6, Column: 2, Message: "エラーの値は errors.go に置く"},
			},
		},
	})
}

func TestCheckSentenceEnd(t *testing.T) {
	t.Parallel()

	runCases(t, []testCase{
		{
			name: "「。」で終わるコメントは違反",
			files: fstest.MapFS{
				"a/check_test.go": file("package a\n\n// 一行目。\n// 途中に。があってもよい\n/*\nブロック。\n*/\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/check_test.go", Line: 3, Column: 1, Message: "コメントを「。」で終えない"},
				{Path: "a/check_test.go", Line: 6, Column: 1, Message: "コメントを「。」で終えない"},
			},
		},
		{
			name: "「。」で終わるエラー文は違反",
			files: fstest.MapFS{
				"a/errors.go": file("package a\n\nimport (\n\t\"errors\"\n\t\"fmt\"\n)\n\n" +
					"var (\n\tErrA = errors.New(\"失敗した。\")\n\tErrB = errors.New(\"失敗した\")\n" +
					"\tErrC = fmt.Errorf(\"失敗した。\")\n)\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/errors.go", Line: 9, Column: 20, Message: "エラー文を「。」で終えない"},
				{Path: "a/errors.go", Line: 11, Column: 20, Message: "エラー文を「。」で終えない"},
			},
		},
		{
			name: "別名で読み込んでも違反",
			files: fstest.MapFS{
				"a/errors.go": file("package a\n\nimport (\n\te \"errors\"\n\tf \"fmt\"\n)\n\n" +
					"var (\n\tErrA = e.New(\"失敗した。\")\n\tErrB = f.Errorf(\"失敗した。\")\n)\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/errors.go", Line: 9, Column: 15, Message: "エラー文を「。」で終えない"},
				{Path: "a/errors.go", Line: 10, Column: 18, Message: "エラー文を「。」で終えない"},
			},
		},
		{
			name: "errors を読み込んでいなければ対象外",
			files: fstest.MapFS{
				"a/check.go": file("package a\n\nfunc f(errors interface{ New(string) error }) error {\n" +
					"\treturn errors.New(\"失敗した。\")\n}\n"),
			},
			want: nil,
		},
	})
}

func TestCheckTestFiles(t *testing.T) {
	t.Parallel()

	runCases(t, []testCase{
		{
			name: "internal のパッケージにテストがあれば通る",
			files: fstest.MapFS{
				"internal/a/check.go":      file("package a\n"),
				"internal/a/check_test.go": file("package a_test\n"),
			},
			want: nil,
		},
		{
			name: "internal のパッケージにテストが無いと違反",
			files: fstest.MapFS{
				"internal/a/model.go":   file("package a\n"),
				"internal/a/check.go":   file("package a\n"),
				"internal/a/b/check.go": file("package b\n"),
			},
			want: []projectlint.Finding{
				{Path: "internal/a/b/check.go", Line: 1, Column: 1, Message: "internal のパッケージには _test.go を置く"},
				{Path: "internal/a/check.go", Line: 1, Column: 1, Message: "internal のパッケージには _test.go を置く"},
			},
		},
		{
			name: "internal の外は対象外",
			files: fstest.MapFS{
				"cmd/a/main.go": file("package main\n"),
			},
			want: nil,
		},
	})
}

func TestCheckIgnoresLineDirectives(t *testing.T) {
	t.Parallel()

	runCases(t, []testCase{
		{
			name: "//line があっても実際のファイルと行で報告する",
			files: fstest.MapFS{
				"a/check.go": file("package a\n\n//line other.go:100\ntype A struct{}\n\n//line other.go:200\n// 説明。\n"),
			},
			want: []projectlint.Finding{
				{Path: "a/check.go", Line: 4, Column: 6, Message: "型の宣言は model.go か *_config.go に置く"},
				{Path: "a/check.go", Line: 7, Column: 1, Message: "コメントを「。」で終えない"},
			},
		},
	})
}

func TestCheckSkipsDirs(t *testing.T) {
	t.Parallel()

	runCases(t, []testCase{
		{
			name: "隠しフォルダと testdata と vendor と Go 以外のファイルは対象外",
			files: fstest.MapFS{
				".git/a.go":       file("package a\n\ntype A struct{}\n"),
				"a/testdata/a.go": file("package a\n\ntype A struct{}\n"),
				"vendor/a/a.go":   file("package a\n\ntype A struct{}\n"),
				"a/README.md":     file("型。\n"),
			},
			want: nil,
		},
		{
			name: "シンボリックリンクは、リンク先があってもなくても対象外",
			files: fstest.MapFS{
				"outside/model.go": file("package outside\n\n// 外。\n"),
				"a/link.go":        {Data: []byte("../outside/model.go"), Mode: fs.ModeSymlink},
				"a/broken.go":      {Data: []byte("../missing.go"), Mode: fs.ModeSymlink},
			},
			want: []projectlint.Finding{
				{Path: "outside/model.go", Line: 3, Column: 1, Message: "コメントを「。」で終えない"},
			},
		},
	})
}

func TestCheckReturnsErrorOnInvalidSource(t *testing.T) {
	t.Parallel()

	_, err := projectlint.Check(fstest.MapFS{"a/check.go": file("package a\n\nfunc {\n")})
	if err == nil {
		t.Fatal("Check() がエラーを返さなかった")
	}
}
