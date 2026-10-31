package audit_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/audit"
	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
)

var (
	errNotFound    = errors.New("見つかりません")
	errWrongTarget = errors.New("取得する版かパッチが違います")
)

type fakeSource struct {
	version   string
	patch     string
	roster    map[int]champion.RiotEntry
	summaries map[int]champion.Summary
	details   map[int]champion.Detail
	games     map[string]champion.GameData
}

func (f fakeSource) LatestVersion(context.Context) (string, error) { return f.version, nil }

func (f fakeSource) Roster(_ context.Context, version string) (map[int]champion.RiotEntry, error) {
	if version != f.version {
		return nil, errWrongTarget
	}

	return f.roster, nil
}

func (f fakeSource) Summaries(_ context.Context, patch string) (map[int]champion.Summary, error) {
	if patch != f.patch {
		return nil, errWrongTarget
	}

	return f.summaries, nil
}

func (f fakeSource) Detail(_ context.Context, patch string, id int) (champion.Detail, error) {
	if patch != f.patch {
		return champion.Detail{}, errWrongTarget
	}

	detail, exists := f.details[id]
	if !exists {
		return champion.Detail{}, errNotFound
	}

	return detail, nil
}

func (f fakeSource) GameData(_ context.Context, patch, alias string) (champion.GameData, error) {
	if patch != f.patch {
		return champion.GameData{}, errWrongTarget
	}

	game, exists := f.games[alias]
	if !exists {
		return champion.GameData{}, errNotFound
	}

	return game, nil
}

func ahriSource() fakeSource {
	spellPaths := []string{
		"Characters/Ahri/Spells/AhriQAbility/AhriQ",
		"Characters/Ahri/Spells/AhriWAbility/AhriW",
		"Characters/Ahri/Spells/AhriEAbility/AhriE",
		"Characters/Ahri/Spells/AhriRAbility/AhriR",
	}
	passivePath := "Characters/Ahri/Spells/AhriPassiveAbility/AhriPassive"

	keys := map[string]struct{}{passivePath: {}}
	for _, path := range spellPaths {
		keys[path] = struct{}{}
	}

	return fakeSource{
		version: "16.19.1",
		patch:   "16.19",
		roster:  map[int]champion.RiotEntry{103: {ID: "Ahri"}},
		summaries: map[int]champion.Summary{
			103:   {ID: 103, Name: "アーリ", Alias: "Ahri"},
			60001: {ID: 60001, Name: "特殊モード", Alias: "Jade_Ahri"},
		},
		details: map[int]champion.Detail{103: {
			ID:      103,
			Passive: champion.Passive{Name: "生気吸引", Description: "体力を回復する"},
			Spells: []champion.Spell{
				{Key: "q", Name: "Q", Description: "Qの説明", DynamicDescription: ""},
				{Key: "w", Name: "W", Description: "Wの説明", DynamicDescription: ""},
				{Key: "e", Name: "E", Description: "Eの説明", DynamicDescription: ""},
				{Key: "r", Name: "R", Description: "Rの説明", DynamicDescription: ""},
			},
		}},
		games: map[string]champion.GameData{"Ahri": {
			Root:       &champion.GameRoot{SpellPaths: spellPaths, PassivePath: passivePath},
			RecordKeys: keys,
		}},
	}
}

func TestRunChecksEveryRiotChampion(t *testing.T) {
	t.Parallel()

	report, err := audit.Run(context.Background(), ahriSource(), "")
	if err != nil {
		t.Fatal(err)
	}

	if report.DataDragonVersion != "16.19.1" || report.Patch != "16.19" {
		t.Fatalf("version = %q, patch = %q", report.DataDragonVersion, report.Patch)
	}

	if report.ChampionCount != 1 || report.StructuralPasses != 1 || len(report.Failures) != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRunReportsMissingSpellRecord(t *testing.T) {
	t.Parallel()

	src := ahriSource()
	delete(src.games["Ahri"].RecordKeys, "Characters/Ahri/Spells/AhriRAbility/AhriR")

	report, err := audit.Run(context.Background(), src, "16.19.1")
	if err != nil {
		t.Fatal(err)
	}

	if report.StructuralPasses != 0 || len(report.Failures) != 1 || !strings.Contains(report.Failures[0].Reason, "ゲーム内スキルレコード") {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRunReportsMissingRootWithDetailProblems(t *testing.T) {
	t.Parallel()

	src := ahriSource()
	src.games["Ahri"] = champion.GameData{Root: nil, RecordKeys: map[string]struct{}{}}
	detail := src.details[103]
	detail.Passive.Description = ""
	src.details[103] = detail

	report, err := audit.Run(context.Background(), src, "16.19.1")
	if err != nil {
		t.Fatal(err)
	}

	if len(report.Failures) != 2 ||
		!strings.Contains(report.Failures[0].Reason, "日本語パッシブ説明") ||
		!strings.Contains(report.Failures[1].Reason, "通常形態のレコード") {
		t.Fatalf("unexpected report: %+v", report.Failures)
	}
}

func TestRunReportsEveryMissingChampion(t *testing.T) {
	t.Parallel()

	src := ahriSource()
	src.roster[266] = champion.RiotEntry{ID: "Aatrox"}

	report, err := audit.Run(context.Background(), src, "16.19.1")
	if err != nil {
		t.Fatal(err)
	}

	if report.ChampionCount != 2 || report.StructuralPasses != 1 || len(report.Failures) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}

	if report.Failures[0].Champion != "Aatrox" || !strings.Contains(report.Failures[0].Reason, "日本語一覧") {
		t.Fatalf("不足チャンピオンを報告していません: %+v", report.Failures)
	}
}
