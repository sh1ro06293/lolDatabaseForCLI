package update

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const (
	installerName = "loldatabase-setup.exe"
	versionParts  = 3
)

func parseVersion(raw string) (version, error) {
	parts := strings.Split(strings.TrimPrefix(raw, "v"), ".")
	if len(parts) != versionParts {
		return version{}, fmt.Errorf("%w: %q", ErrInvalidVersion, raw)
	}

	var parsed version

	for i, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return version{}, fmt.Errorf("%w: %q", ErrInvalidVersion, raw)
		}

		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return version{}, fmt.Errorf("%w: %q: %w", ErrInvalidVersion, raw, err)
		}

		parsed[i] = n
	}

	return parsed, nil
}

func (v version) less(other version) bool {
	for i := range v {
		if v[i] != other[i] {
			return v[i] < other[i]
		}
	}

	return false
}

// Check は明確に必須更新と確認できた場合だけ UpdateRequired を true にする
// 更新情報が取得できない場合は停止判定にせず、エラーを Status に残す
func Check(ctx context.Context, src Source, currentVersion string) Status {
	var result Status

	current, err := parseVersion(currentVersion)
	if err != nil {
		result.ReleaseError = err
		return result
	}

	latest, err := src.LatestRelease(ctx)
	if err != nil {
		result.ReleaseError = err
		return result
	}

	latestVersion, err := parseVersion(latest.Tag)
	if err != nil {
		result.ReleaseError = err
		return result
	}

	result.LatestVersion = latest.Tag

	result.ReleaseURL = latest.URL
	for _, asset := range latest.Assets {
		if asset.Name == installerName {
			result.InstallerURL = asset.URL
			break
		}
	}

	if result.InstallerURL == "" {
		result.ReleaseError = ErrInstallerNotFound
		return result
	}

	result.UpdateAvailable = current.less(latestVersion)

	rules, err := src.Policy(ctx)
	if err != nil {
		result.PolicyError = err
		return result
	}

	minimum, err := parseVersion(rules.MinimumSupportedVersion)
	if err != nil {
		result.PolicyError = err
		return result
	}

	if latestVersion.less(minimum) {
		result.PolicyError = ErrMinimumVersionUnreleased
		return result
	}

	result.UpdateRequired = current.less(minimum)

	return result
}
