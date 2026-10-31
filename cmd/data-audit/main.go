package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/audit"
	"github.com/sh1ro06293/lolDatabaseForCLI/internal/riot"
)

const (
	auditTimeout = 15 * time.Minute
	exitOK       = 0
	exitFailures = 1
	exitError    = 2
)

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}

func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("data-audit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	version := flags.String("version", "", "Data Dragon のバージョン。省略時は最新版")

	err := flags.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}

	if err != nil {
		return exitError
	}

	ctx, cancel := context.WithTimeout(context.Background(), auditTimeout)
	defer cancel()

	client := riot.Client{
		HTTPClient:         &http.Client{Timeout: riot.DefaultTimeout},
		DataDragonURL:      riot.DefaultDataDragonURL,
		CommunityDragonURL: riot.DefaultCommunityDragonURL,
	}

	report, err := audit.Run(ctx, client, *version)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "データ検証を開始できません:", err)
		return exitError
	}

	err = json.NewEncoder(stdout).Encode(report)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "検証結果を出力できません:", err)
		return exitError
	}

	if len(report.Failures) > 0 {
		return exitFailures
	}

	return exitOK
}
