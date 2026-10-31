package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	policycheckcmd "github.com/sh1ro06293/lolDatabaseForCLI/cmd/policycheck"
)

func writePolicy(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func policyJSON(minimum string) string {
	return `{"minimumSupportedVersion": "` + minimum + `"}`
}

func runWithPolicies(t *testing.T, before, after string) (int, string, string) {
	t.Helper()

	dir := t.TempDir()
	beforePath := writePolicy(t, dir, "before.json", before)
	afterPath := writePolicy(t, dir, "after.json", after)

	var stdout, stderr bytes.Buffer

	code := policycheckcmd.Run([]string{beforePath, afterPath}, &stdout, &stderr)

	return code, stdout.String(), stderr.String()
}

func TestRunAccepts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		before     string
		after      string
		wantStdout string
	}{
		{
			name: "同じ値なら 0", before: policyJSON("v1.2.0"), after: policyJSON("v1.2.0"),
			wantStdout: "最低利用可能バージョンは下がっていません: v1.2.0 から v1.2.0\n",
		},
		{
			name: "上げれば 0", before: policyJSON("v1.2.0"), after: policyJSON("v1.3.0"),
			wantStdout: "最低利用可能バージョンは下がっていません: v1.2.0 から v1.3.0\n",
		},
		{
			name: "変更前が JSON として読めなくても、変更後が正しければ 0", before: `{`, after: policyJSON("v1.0.0"),
			wantStdout: "変更前の最低利用可能バージョンを読めないため比べず、変更後の値だけを確かめました: v1.0.0\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, stdout, stderr := runWithPolicies(t, tt.before, tt.after)
			if code != 0 {
				t.Fatalf("Run() = %d, want 0 (stderr: %s)", code, stderr)
			}

			if stdout != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout, tt.wantStdout)
			}

			if stderr != "" {
				t.Fatalf("stderr = %q, want 空", stderr)
			}
		})
	}
}

func TestRunRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		before     string
		after      string
		wantCode   int
		wantStderr string
	}{
		{name: "下げれば 1", before: policyJSON("v1.2.0"), after: policyJSON("v1.1.0"), wantCode: 1, wantStderr: "最低利用可能バージョンが下がっています"},
		{name: "不正な値なら 1", before: policyJSON("v1.2.0"), after: policyJSON("v1.0"), wantCode: 1, wantStderr: "バージョン形式が不正です"},
		{name: "項目が無ければ不正な値として 1", before: policyJSON("v1.2.0"), after: `{}`, wantCode: 1, wantStderr: "バージョン形式が不正です"},
		{name: "変更後が JSON として読めなければ 2", before: policyJSON("v1.2.0"), after: `{`, wantCode: 2, wantStderr: "JSON として読めません"},
		{
			name: "変更前が JSON として読めなくても、変更後が不正なら 1", before: `{`, after: policyJSON("v1.0"),
			wantCode: 1, wantStderr: "バージョン形式が不正です",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, stdout, stderr := runWithPolicies(t, tt.before, tt.after)
			if code != tt.wantCode {
				t.Fatalf("Run() = %d, want %d (stderr: %s)", code, tt.wantCode, stderr)
			}

			if !strings.Contains(stderr, tt.wantStderr) {
				t.Fatalf("stderr = %q, want %q を含む", stderr, tt.wantStderr)
			}

			if stdout != "" {
				t.Fatalf("stdout = %q, want 空", stdout)
			}
		})
	}
}

func TestRunReportsUsageAndUnreadableFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existing := writePolicy(t, dir, "policy.json", policyJSON("v1.0.0"))
	missing := filepath.Join(dir, "missing.json")

	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{name: "引数が足りなければ 2", args: []string{existing}, wantStderr: "使い方: policycheck"},
		{name: "引数が多ければ 2", args: []string{existing, existing, existing}, wantStderr: "使い方: policycheck"},
		{name: "変更前が無ければ 2", args: []string{missing, existing}, wantStderr: "を読めません"},
		{name: "変更後が無ければ 2", args: []string{existing, missing}, wantStderr: "を読めません"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := policycheckcmd.Run(tt.args, &stdout, &stderr)
			if code != 2 {
				t.Fatalf("Run() = %d, want 2", code)
			}

			if !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q, want %q を含む", stderr.String(), tt.wantStderr)
			}
		})
	}
}
