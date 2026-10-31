package github_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/github"
	"github.com/sh1ro06293/lolDatabaseForCLI/internal/update"
)

var _ update.Source = (*github.Client)(nil)

func testClient(t *testing.T, policyStatus int) github.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/latest":
			_, _ = fmt.Fprint(w, `{"tag_name":"v1.2.0","html_url":"https://github.com/example/releases/latest",`+
				`"assets":[{"name":"loldatabase-setup.exe","browser_download_url":"https://github.com/example/setup.exe"}]}`)
		case "/policy":
			if policyStatus != http.StatusOK {
				w.WriteHeader(policyStatus)
				return
			}

			_, _ = fmt.Fprint(w, `{"minimumSupportedVersion":"v1.0.0"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return github.Client{HTTPClient: server.Client(), LatestReleaseURL: server.URL + "/latest", PolicyURL: server.URL + "/policy"}
}

func TestLatestRelease(t *testing.T) {
	t.Parallel()

	release, err := testClient(t, http.StatusOK).LatestRelease(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	want := update.Release{
		Tag:    "v1.2.0",
		URL:    "https://github.com/example/releases/latest",
		Assets: []update.Asset{{Name: "loldatabase-setup.exe", URL: "https://github.com/example/setup.exe"}},
	}
	if release.Tag != want.Tag || release.URL != want.URL || !slices.Equal(release.Assets, want.Assets) {
		t.Fatalf("LatestRelease() = %+v, want %+v", release, want)
	}
}

func TestPolicy(t *testing.T) {
	t.Parallel()

	policy, err := testClient(t, http.StatusOK).Policy(context.Background())
	if err != nil || policy.MinimumSupportedVersion != "v1.0.0" {
		t.Fatalf("Policy() = %+v, %v", policy, err)
	}
}

func TestPolicyReportsHTTPStatus(t *testing.T) {
	t.Parallel()

	_, err := testClient(t, http.StatusServiceUnavailable).Policy(context.Background())
	if !errors.Is(err, github.ErrUnexpectedStatus) {
		t.Fatal("HTTP 503 でエラーになりませんでした")
	}
}
