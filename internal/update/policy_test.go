package update_test

import (
	"errors"
	"testing"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/update"
)

func TestCheckMinimumNotLowered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		before       string
		after        string
		wantCompared bool
		wantErr      error
	}{
		{name: "同じ値は通る", before: "v1.2.0", after: "v1.2.0", wantCompared: true, wantErr: nil},
		{name: "上げる変更は通る", before: "v1.2.0", after: "v1.3.0", wantCompared: true, wantErr: nil},
		{name: "パッチ版だけ上げても通る", before: "v1.2.0", after: "v1.2.1", wantCompared: true, wantErr: nil},
		{name: "桁数が増えても数値で比べる", before: "v1.9.0", after: "v1.10.0", wantCompared: true, wantErr: nil},
		{name: "下げる変更は拒否する", before: "v1.2.0", after: "v1.1.9", wantCompared: true, wantErr: update.ErrMinimumLowered},
		{name: "桁数が増えても数値で比べて下げを拒否する", before: "v1.10.0", after: "v1.9.0", wantCompared: true, wantErr: update.ErrMinimumLowered},
		{name: "v 無しの値も比べられる", before: "1.2.0", after: "1.1.0", wantCompared: true, wantErr: update.ErrMinimumLowered},
		{name: "変更後が版として不正なら拒否する", before: "v1.2.0", after: "v1.0", wantCompared: false, wantErr: update.ErrInvalidVersion},
		{name: "変更後が空なら拒否する", before: "v1.2.0", after: "", wantCompared: false, wantErr: update.ErrInvalidVersion},
		{name: "変更前が不正でも変更後が正しければ通る", before: "v1.0", after: "v0.0.0", wantCompared: false, wantErr: nil},
		{name: "変更前も変更後も不正なら拒否する", before: "v1.0", after: "v2.0", wantCompared: false, wantErr: update.ErrInvalidVersion},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			compared, err := update.CheckMinimumNotLowered(tt.before, tt.after)
			if compared != tt.wantCompared {
				t.Fatalf("CheckMinimumNotLowered(%q, %q) の比べたかどうか = %t, want %t", tt.before, tt.after, compared, tt.wantCompared)
			}

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("CheckMinimumNotLowered(%q, %q) = %v, want nil", tt.before, tt.after, err)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckMinimumNotLowered(%q, %q) = %v, want %v", tt.before, tt.after, err, tt.wantErr)
			}
		})
	}
}
