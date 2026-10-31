package github

import (
	"net/http"
	"time"
)

const (
	DefaultLatestReleaseURL = "https://api.github.com/repos/sh1ro06293/lolDatabaseForCLI/releases/latest"
	DefaultPolicyURL        = "https://raw.githubusercontent.com/sh1ro06293/lolDatabaseForCLI/main/update-policy.json"
	DefaultTimeout          = 5 * time.Second
)

// Client は CLI 本体の更新情報を公開リポジトリから取得する
type Client struct {
	HTTPClient       *http.Client
	LatestReleaseURL string
	PolicyURL        string
}
