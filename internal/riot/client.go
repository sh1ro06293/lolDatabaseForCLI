package riot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
)

const maxResponseBytes = 32 << 20

func (c Client) getJSON(ctx context.Context, base, path string, dst any) error {
	endpoint, err := url.JoinPath(base, path)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRequestFailed, endpoint, err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %s: HTTP %d", ErrUnexpectedStatus, endpoint, resp.StatusCode)
	}

	decoder := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes))

	err = decoder.Decode(dst)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrInvalidResponse, endpoint, err)
	}

	return nil
}

func (c Client) LatestVersion(ctx context.Context) (string, error) {
	var versions []string

	err := c.getJSON(ctx, c.DataDragonURL, "api/versions.json", &versions)
	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		return "", ErrEmptyVersions
	}

	return versions[0], nil
}

func (c Client) Roster(ctx context.Context, version string) (map[int]champion.RiotEntry, error) {
	var response rosterResponse

	path := "cdn/" + version + "/data/ja_JP/champion.json"

	err := c.getJSON(ctx, c.DataDragonURL, path, &response)
	if err != nil {
		return nil, err
	}

	if len(response.Data) == 0 {
		return nil, ErrEmptyRoster
	}

	roster := make(map[int]champion.RiotEntry, len(response.Data))
	for _, entry := range response.Data {
		id, err := strconv.Atoi(entry.Key)
		if err != nil || id <= 0 || entry.ID == "" {
			return nil, fmt.Errorf("%w: %+v", ErrInvalidChampionID, entry)
		}

		if _, exists := roster[id]; exists {
			return nil, fmt.Errorf("%w: %d", ErrDuplicateChampionID, id)
		}

		roster[id] = champion.RiotEntry{ID: entry.ID}
	}

	return roster, nil
}

func (c Client) Summaries(ctx context.Context, patch string) (map[int]champion.Summary, error) {
	path := patch + "/plugins/rcp-be-lol-game-data/global/ja_jp/v1/champion-summary.json"

	var all []summary

	err := c.getJSON(ctx, c.CommunityDragonURL, path, &all)
	if err != nil {
		return nil, err
	}

	result := make(map[int]champion.Summary, len(all))
	for _, s := range all {
		if s.ID > 0 {
			result[s.ID] = champion.Summary{ID: s.ID, Name: s.Name, Alias: s.Alias}
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnexpectedFormat, path)
	}

	return result, nil
}

func (c Client) Detail(ctx context.Context, patch string, id int) (champion.Detail, error) {
	path := fmt.Sprintf("%s/plugins/rcp-be-lol-game-data/global/ja_jp/v1/champions/%d.json", patch, id)

	var raw detail

	err := c.getJSON(ctx, c.CommunityDragonURL, path, &raw)
	if err != nil {
		return champion.Detail{}, err
	}

	if raw.ID <= 0 {
		return champion.Detail{}, fmt.Errorf("%w: %s", ErrUnexpectedFormat, path)
	}

	return toDetail(raw), nil
}

func (c Client) GameData(ctx context.Context, patch, alias string) (champion.GameData, error) {
	name := strings.ToLower(alias)
	path := fmt.Sprintf("%s/game/data/characters/%s/%s.bin.json", patch, name, name)

	var records map[string]json.RawMessage

	err := c.getJSON(ctx, c.CommunityDragonURL, path, &records)
	if err != nil {
		return champion.GameData{}, err
	}

	keys := make(map[string]struct{}, len(records))
	for key := range records {
		keys[key] = struct{}{}
	}

	game := champion.GameData{Root: nil, RecordKeys: keys}
	// 見つからないときはエラーにせず Root を nil のままにする。エラーにすると、監査が日本語の説明の欠落を報告しなくなる
	raw, found := findRecord(records, "Characters/"+alias+"/CharacterRecords/Root")
	if !found {
		return game, nil
	}

	var root gameRoot

	err = json.Unmarshal(raw, &root)
	if err != nil {
		return champion.GameData{}, fmt.Errorf("%w: %s: %w", ErrInvalidResponse, path, err)
	}

	game.Root = &champion.GameRoot{SpellPaths: root.Spells, PassivePath: root.Passive}

	return game, nil
}
