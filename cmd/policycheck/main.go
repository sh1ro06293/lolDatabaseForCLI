package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/update"
)

const (
	exitOK       = 0
	exitRejected = 1
	exitError    = 2

	argCount = 2
)

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) != argCount {
		_, _ = fmt.Fprintln(stderr, ErrUsage)

		return exitError
	}

	beforeData, err := readFile(args[0])
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)

		return exitError
	}

	afterData, err := readFile(args[1])
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)

		return exitError
	}

	after, err := decodeMinimum(args[1], afterData)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)

		return exitError
	}

	// 変更前が JSON として読めなくても止めない。止めると、壊れた値を直す変更まで拒否して修復できなくなる
	before, _ := decodeMinimum(args[0], beforeData)

	compared, err := update.CheckMinimumNotLowered(before, after)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "update-policy.json: %v\n", err)

		if errors.Is(err, update.ErrMinimumLowered) || errors.Is(err, update.ErrInvalidVersion) {
			return exitRejected
		}

		return exitError
	}

	if !compared {
		_, _ = fmt.Fprintf(stdout, "変更前の最低利用可能バージョンを読めないため比べず、変更後の値だけを確かめました: %s\n", after)

		return exitOK
	}

	_, _ = fmt.Fprintf(stdout, "最低利用可能バージョンは下がっていません: %s から %s\n", before, after)

	return exitOK
}

func readFile(path string) ([]byte, error) {
	data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("%s を読めません: %w", path, err)
	}

	return data, nil
}

func decodeMinimum(path string, data []byte) (string, error) {
	var policy policyFile

	err := json.Unmarshal(data, &policy)
	if err != nil {
		return "", fmt.Errorf("%s を JSON として読めません: %w", path, err)
	}

	return policy.MinimumSupportedVersion, nil
}
