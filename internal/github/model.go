package github

type release struct {
	TagName string  `json:"tag_name"` //nolint:tagliatelle // GitHub API の項目名
	HTMLURL string  `json:"html_url"` //nolint:tagliatelle // GitHub API の項目名
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"` //nolint:tagliatelle // GitHub API の項目名
}

type policy struct {
	MinimumSupportedVersion string `json:"minimumSupportedVersion"`
}
