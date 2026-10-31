package riot

import (
	"net/http"
	"time"
)

const (
	DefaultDataDragonURL      = "https://ddragon.leagueoflegends.com"
	DefaultCommunityDragonURL = "https://raw.communitydragon.org"
	DefaultTimeout            = 20 * time.Second
)

// Client は同一パッチの Riot / CommunityDragon データを取得する
// URL と HTTPClient はテスト時に差し替えられる
type Client struct {
	HTTPClient         *http.Client
	DataDragonURL      string
	CommunityDragonURL string
}
