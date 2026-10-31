package riot_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
	"github.com/sh1ro06293/lolDatabaseForCLI/internal/riot"
)

func testClient(t *testing.T, responses map[string]any) riot.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response, exists := responses[r.URL.Path]
		if !exists {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	return riot.Client{HTTPClient: server.Client(), DataDragonURL: server.URL, CommunityDragonURL: server.URL}
}

func TestLatestVersion(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{"/api/versions.json": []string{"16.19.1", "16.18.1"}})

	version, err := client.LatestVersion(context.Background())
	if err != nil || version != "16.19.1" {
		t.Fatalf("LatestVersion() = %q, %v", version, err)
	}
}

func TestRoster(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{
		"/cdn/16.19.1/data/ja_JP/champion.json": map[string]any{
			"data": map[string]any{"Ahri": map[string]any{"id": "Ahri", "key": "103"}},
		},
	})

	roster, err := client.Roster(context.Background(), "16.19.1")
	if err != nil {
		t.Fatal(err)
	}

	if len(roster) != 1 || roster[103] != (champion.RiotEntry{ID: "Ahri"}) {
		t.Fatalf("Roster() = %+v", roster)
	}
}

func TestSummariesSkipsInvalidID(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{
		"/16.19/plugins/rcp-be-lol-game-data/global/ja_jp/v1/champion-summary.json": []any{
			map[string]any{"id": 103, "name": "アーリ", "alias": "Ahri"},
			map[string]any{"id": -1, "name": "なし", "alias": "None"},
		},
	})

	summaries, err := client.Summaries(context.Background(), "16.19")
	if err != nil {
		t.Fatal(err)
	}

	if len(summaries) != 1 || summaries[103] != (champion.Summary{ID: 103, Name: "アーリ", Alias: "Ahri"}) {
		t.Fatalf("Summaries() = %+v", summaries)
	}
}

func TestSummariesReportsUnexpectedFormat(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{
		"/16.19/plugins/rcp-be-lol-game-data/global/ja_jp/v1/champion-summary.json": []any{
			map[string]any{"championId": 103, "name": "アーリ", "alias": "Ahri"},
		},
	})

	_, err := client.Summaries(context.Background(), "16.19")
	if !errors.Is(err, riot.ErrUnexpectedFormat) {
		t.Fatalf("Summaries() error = %v, want ErrUnexpectedFormat", err)
	}
}

func TestDetail(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{
		"/16.19/plugins/rcp-be-lol-game-data/global/ja_jp/v1/champions/103.json": map[string]any{
			"id":      103,
			"passive": map[string]any{"name": "生気吸引", "description": "体力を回復する"},
			"spells": []any{
				map[string]any{"spellKey": "q", "name": "Q", "description": "Qの説明", "dynamicDescription": "Qの動的説明"},
			},
		},
	})

	detail, err := client.Detail(context.Background(), "16.19", 103)
	if err != nil {
		t.Fatal(err)
	}

	want := champion.Detail{
		ID:      103,
		Passive: champion.Passive{Name: "生気吸引", Description: "体力を回復する"},
		Spells:  []champion.Spell{{Key: "q", Name: "Q", Description: "Qの説明", DynamicDescription: "Qの動的説明"}},
	}
	if detail.ID != want.ID || detail.Passive != want.Passive || !slices.Equal(detail.Spells, want.Spells) {
		t.Fatalf("Detail() = %+v, want %+v", detail, want)
	}
}

func TestDetailReportsUnexpectedFormat(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{
		"/16.19/plugins/rcp-be-lol-game-data/global/ja_jp/v1/champions/103.json": map[string]any{
			"championId": 103,
			"passive":    map[string]any{"name": "生気吸引", "description": "体力を回復する"},
		},
	})

	_, err := client.Detail(context.Background(), "16.19", 103)
	if !errors.Is(err, riot.ErrUnexpectedFormat) {
		t.Fatalf("Detail() error = %v, want ErrUnexpectedFormat", err)
	}
}

func TestGameData(t *testing.T) {
	t.Parallel()

	const path = "/16.19/game/data/characters/ahri/ahri.bin.json"

	tests := []struct {
		name     string
		records  map[string]any
		wantRoot *champion.GameRoot
	}{
		{
			name: "通常形態のレコードは大文字と小文字を区別せずに探す",
			records: map[string]any{
				"characters/ahri/characterrecords/root": map[string]any{
					"spells":                 []string{"Characters/Ahri/Spells/AhriQ"},
					"mCharacterPassiveSpell": "Characters/Ahri/Spells/AhriPassive",
				},
				"Characters/Ahri/Spells/AhriQ": map[string]any{},
			},
			wantRoot: &champion.GameRoot{
				SpellPaths:  []string{"Characters/Ahri/Spells/AhriQ"},
				PassivePath: "Characters/Ahri/Spells/AhriPassive",
			},
		},
		{
			name:     "通常形態のレコードが無ければ Root は nil",
			records:  map[string]any{"Characters/Ahri/Spells/AhriQ": map[string]any{}},
			wantRoot: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := testClient(t, map[string]any{path: tt.records})

			game, err := client.GameData(context.Background(), "16.19", "Ahri")
			if err != nil {
				t.Fatal(err)
			}

			if (game.Root == nil) != (tt.wantRoot == nil) ||
				(game.Root != nil && (game.Root.PassivePath != tt.wantRoot.PassivePath || !slices.Equal(game.Root.SpellPaths, tt.wantRoot.SpellPaths))) {
				t.Fatalf("Root = %+v, want %+v", game.Root, tt.wantRoot)
			}

			if len(game.RecordKeys) != len(tt.records) {
				t.Fatalf("RecordKeys = %v", game.RecordKeys)
			}

			for key := range tt.records {
				if _, exists := game.RecordKeys[key]; !exists {
					t.Fatalf("RecordKeys に %q がありません", key)
				}
			}
		})
	}
}

func TestGameDataReportsUnreadableRoot(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{
		"/16.19/game/data/characters/ahri/ahri.bin.json": map[string]any{
			"Characters/Ahri/CharacterRecords/Root": "オブジェクトではない",
		},
	})

	_, err := client.GameData(context.Background(), "16.19", "Ahri")
	if !errors.Is(err, riot.ErrInvalidResponse) {
		t.Fatal("通常形態のレコードが読めないのにエラーになりませんでした")
	}
}

func TestGetJSONReportsHTTPStatus(t *testing.T) {
	t.Parallel()

	client := testClient(t, map[string]any{})

	_, err := client.LatestVersion(context.Background())
	if !errors.Is(err, riot.ErrUnexpectedStatus) {
		t.Fatal("HTTP 404 でエラーになりませんでした")
	}
}
