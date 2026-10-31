package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/projectlint"
)

const (
	exitOK       = 0
	exitFindings = 1
	exitError    = 2
)

func main() {
	os.Exit(Run(os.DirFS("."), os.Stdout, os.Stderr))
}

func Run(fsys fs.FS, stdout, stderr io.Writer) int {
	findings, err := projectlint.Check(fsys)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)

		return exitError
	}

	for _, f := range findings {
		_, _ = fmt.Fprintf(stdout, "%s:%d:%d: %s (projectlint)\n", f.Path, f.Line, f.Column, f.Message)
	}

	if len(findings) > 0 {
		_, _ = fmt.Fprintf(stdout, "%d 件\n", len(findings))

		return exitFindings
	}

	return exitOK
}
