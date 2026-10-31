package champion

import (
	"fmt"
	"strconv"
	"strings"
)

// パッチは版の先頭 2 つ（16.20.1 なら 16.20）
const patchParts = 2

func PatchFromVersion(version string) (string, error) {
	parts := strings.Split(version, ".")
	if len(parts) < patchParts {
		return "", fmt.Errorf("%w: %q", ErrInvalidVersion, version)
	}

	for _, part := range parts[:patchParts] {
		_, err := strconv.Atoi(part)
		if err != nil {
			return "", fmt.Errorf("%w: %q: %w", ErrInvalidVersion, version, err)
		}
	}

	return strings.Join(parts[:patchParts], "."), nil
}
