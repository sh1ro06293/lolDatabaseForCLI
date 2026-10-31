package main_test

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	projectlintcmd "github.com/sh1ro06293/lolDatabaseForCLI/cmd/projectlint"
)

func TestRunExitCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		files      fstest.MapFS
		wantCode   int
		wantStdout string
	}{
		{
			name: "違反が無ければ 0",
			files: fstest.MapFS{
				"internal/a/check.go":      {Data: []byte("package a\n")},
				"internal/a/check_test.go": {Data: []byte("package a_test\n")},
			},
			wantCode:   0,
			wantStdout: "",
		},
		{
			name:       "違反があれば指摘と件数を出して 1",
			files:      fstest.MapFS{"internal/a/check.go": {Data: []byte("package a\n")}},
			wantCode:   1,
			wantStdout: "internal/a/check.go:1:1: internal のパッケージには _test.go を置く (projectlint)\n1 件\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := projectlintcmd.Run(tt.files, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("Run() = %d, want %d (stderr: %s)", code, tt.wantCode, stderr.String())
			}

			if stdout.String() != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout.String(), tt.wantStdout)
			}

			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want 空", stderr.String())
			}
		})
	}
}

func TestRunReportsUnreadableSource(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := projectlintcmd.Run(fstest.MapFS{"internal/a/check.go": {Data: []byte("package a\n\nfunc {\n")}}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run() = %d, want 2", code)
	}

	// 後ろに続く構文エラーの文面は Go の版で変わるので、前置きの部分だけを見る
	if !strings.HasPrefix(stderr.String(), "internal/a/check.go を解析できません: ") {
		t.Fatalf("stderr = %q, want 「internal/a/check.go を解析できません: 」で始まる", stderr.String())
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want 空", stdout.String())
	}
}
