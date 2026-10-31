package update_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/update"
)

var errUnavailable = errors.New("取得できません")

type fakeSource struct {
	release    update.Release
	releaseErr error
	policy     update.Policy
	policyErr  error
}

func (f fakeSource) LatestRelease(context.Context) (update.Release, error) {
	return f.release, f.releaseErr
}

func (f fakeSource) Policy(context.Context) (update.Policy, error) {
	return f.policy, f.policyErr
}

func newSource(latest, minimum string, withInstaller bool) fakeSource {
	var assets []update.Asset
	if withInstaller {
		assets = []update.Asset{{Name: "loldatabase-setup.exe", URL: "https://github.com/example/setup.exe"}}
	}

	return fakeSource{
		release:    update.Release{Tag: latest, URL: "https://github.com/example/releases/latest", Assets: assets},
		releaseErr: nil,
		policy:     update.Policy{MinimumSupportedVersion: minimum},
		policyErr:  nil,
	}
}

func isError(got, want error) bool {
	if want == nil {
		return got == nil
	}

	return errors.Is(got, want)
}

func TestCheck(t *testing.T) {
	t.Parallel()

	policyUnavailable := newSource("v1.2.0", "", true)
	policyUnavailable.policyErr = errUnavailable

	tests := []struct {
		name           string
		current        string
		src            fakeSource
		wantAvailable  bool
		wantRequired   bool
		wantPolicyErr  error
		wantReleaseErr error
	}{
		{name: "最新版", current: "v1.2.0", src: newSource("v1.2.0", "v1.0.0", true)},
		{name: "任意更新", current: "v1.1.0", src: newSource("v1.2.0", "v1.0.0", true), wantAvailable: true},
		{name: "必須の版を飛ばしたまま任意の版が出ても必須更新になる", current: "v1.0.0", src: newSource("v1.2.0", "v1.1.0", true), wantAvailable: true, wantRequired: true},
		{name: "更新方針を取得できない", current: "v1.0.0", src: policyUnavailable, wantAvailable: true, wantPolicyErr: errUnavailable},
		{name: "公開前の最低バージョン", current: "v1.0.0", src: newSource("v1.2.0", "v1.3.0", true), wantAvailable: true, wantPolicyErr: update.ErrMinimumVersionUnreleased},
		{name: "インストーラーがない", current: "v1.0.0", src: newSource("v1.2.0", "v1.1.0", false), wantReleaseErr: update.ErrInstallerNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status := update.Check(context.Background(), tt.src, tt.current)
			if status.UpdateAvailable != tt.wantAvailable || status.UpdateRequired != tt.wantRequired {
				t.Fatalf("更新判定: got available=%v required=%v, want available=%v required=%v", status.UpdateAvailable, status.UpdateRequired, tt.wantAvailable, tt.wantRequired)
			}

			if !isError(status.PolicyError, tt.wantPolicyErr) || !isError(status.ReleaseError, tt.wantReleaseErr) {
				t.Fatalf("エラー判定: policy=%v release=%v", status.PolicyError, status.ReleaseError)
			}
		})
	}
}

func TestCheckReleaseUnavailableDoesNotBlock(t *testing.T) {
	t.Parallel()

	src := newSource("v1.2.0", "v1.0.0", true)
	src.releaseErr = errUnavailable
	src.policyErr = errUnavailable

	status := update.Check(context.Background(), src, "v1.0.0")
	if status.UpdateRequired || !errors.Is(status.ReleaseError, errUnavailable) {
		t.Fatalf("取得失敗時は停止しない: %+v", status)
	}
}

func TestCheckCurrentVersionFormat(t *testing.T) {
	t.Parallel()

	src := newSource("v1.2.3", "v0.0.0", true)
	for _, value := range []string{"v1.2.3", "16.20.1", "v0.0.0"} {
		if status := update.Check(context.Background(), src, value); status.ReleaseError != nil {
			t.Errorf("%q: %v", value, status.ReleaseError)
		}
	}

	for _, value := range []string{"", "dev", "v1.2", "v1.2.3-beta", "v01.2.3", "v1.-2.3"} {
		if status := update.Check(context.Background(), src, value); !errors.Is(status.ReleaseError, update.ErrInvalidVersion) {
			t.Errorf("不正なバージョンを受け入れました: %q", value)
		}
	}
}
