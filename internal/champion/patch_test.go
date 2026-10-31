package champion_test

import (
	"errors"
	"testing"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
)

func TestPatchFromVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version string
		patch   string
		wantErr error
	}{
		{version: "16.19.1", patch: "16.19"},
		{version: "16.20", patch: "16.20"},
		{version: "invalid", wantErr: champion.ErrInvalidVersion},
		{version: "16.x.1", wantErr: champion.ErrInvalidVersion},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			t.Parallel()

			patch, err := champion.PatchFromVersion(tt.version)
			if !errors.Is(err, tt.wantErr) || patch != tt.patch {
				t.Fatalf("PatchFromVersion(%q) = %q, %v", tt.version, patch, err)
			}
		})
	}
}
