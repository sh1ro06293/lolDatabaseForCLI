package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/update"
)

const maxResponseBytes = 1 << 20

func (c Client) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRequestFailed, url, err)
	}

	req.Header.Set("User-Agent", "loldatabase")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s: HTTP %d", ErrUnexpectedStatus, url, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrInvalidResponse, url, err)
	}

	if len(body) > maxResponseBytes {
		return fmt.Errorf("%w: %s", ErrResponseTooLarge, url)
	}

	err = json.Unmarshal(body, dst)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrInvalidResponse, url, err)
	}

	return nil
}

func (c Client) LatestRelease(ctx context.Context) (update.Release, error) {
	var latest release

	err := c.getJSON(ctx, c.LatestReleaseURL, &latest)
	if err != nil {
		return update.Release{}, err
	}

	assets := make([]update.Asset, 0, len(latest.Assets))
	for _, a := range latest.Assets {
		assets = append(assets, update.Asset{Name: a.Name, URL: a.BrowserDownloadURL})
	}

	return update.Release{Tag: latest.TagName, URL: latest.HTMLURL, Assets: assets}, nil
}

func (c Client) Policy(ctx context.Context) (update.Policy, error) {
	var raw policy

	err := c.getJSON(ctx, c.PolicyURL, &raw)
	if err != nil {
		return update.Policy{}, err
	}

	return update.Policy{MinimumSupportedVersion: raw.MinimumSupportedVersion}, nil
}
