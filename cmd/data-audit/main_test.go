package main_test

import (
	"bytes"
	"strings"
	"testing"

	dataaudit "github.com/sh1ro06293/lolDatabaseForCLI/cmd/data-audit"
)

func TestRunExitCodesWithoutNetwork(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string
	}{
		{name: "-h は使い方を出して 0", args: []string{"-h"}, wantCode: 0, wantStderr: "Usage of data-audit"},
		{name: "知らないフラグは 2", args: []string{"-nope"}, wantCode: 2, wantStderr: "flag provided but not defined"},
		{name: "不正な版は通信する前に 2", args: []string{"-version", "x.y"}, wantCode: 2, wantStderr: "データ検証を開始できません"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := dataaudit.Run(tt.args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("Run() = %d, want %d (stderr: %s)", code, tt.wantCode, stderr.String())
			}

			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q, want %q を含む", stderr.String(), tt.wantStderr)
			}

			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want 空", stdout.String())
			}
		})
	}
}
