package update

import "context"

// Source は更新情報の取得元
type Source interface {
	LatestRelease(ctx context.Context) (Release, error)
	Policy(ctx context.Context) (Policy, error)
}

type Status struct {
	LatestVersion   string
	ReleaseURL      string
	InstallerURL    string
	UpdateAvailable bool
	UpdateRequired  bool
	ReleaseError    error
	PolicyError     error
}

type Release struct {
	Tag    string
	URL    string
	Assets []Asset
}

type Asset struct {
	Name string
	URL  string
}

type Policy struct {
	MinimumSupportedVersion string
}

type version [versionParts]uint64
