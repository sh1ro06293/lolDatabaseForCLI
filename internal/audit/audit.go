package audit

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
)

// Q・W・E・R の 4 つ
const spellCount = 4

func hasRecord(records map[string]struct{}, path string) bool {
	if _, exists := records[path]; exists {
		return true
	}

	for key := range records {
		if strings.EqualFold(key, path) {
			return true
		}
	}

	return false
}

func checkDetail(detail champion.Detail) []string {
	var problems []string
	if detail.Passive.Name == "" || detail.Passive.Description == "" {
		problems = append(problems, "日本語パッシブ説明がありません")
	}

	if len(detail.Spells) < spellCount {
		problems = append(problems, fmt.Sprintf("日本語スキルが %d 件しかありません", len(detail.Spells)))
	}

	seenSpells := make(map[string]bool, len(detail.Spells))
	for _, spell := range detail.Spells {
		seenSpells[strings.ToLower(spell.Key)] = true
		if spell.Name == "" || (spell.Description == "" && spell.DynamicDescription == "") {
			problems = append(problems, "日本語スキル説明がありません: "+spell.Key)
		}
	}

	for _, key := range []string{"q", "w", "e", "r"} {
		if !seenSpells[key] {
			problems = append(problems, "日本語スキルがありません: "+strings.ToUpper(key))
		}
	}

	return problems
}

func checkGameData(game champion.GameData, alias string) []string {
	root := game.Root
	if root == nil {
		return []string{"通常形態のレコードがありません: " + alias}
	}

	var problems []string
	if len(root.SpellPaths) < spellCount {
		problems = append(problems, fmt.Sprintf("ゲーム内スキル参照が %d 件しかありません", len(root.SpellPaths)))
	}

	if root.PassivePath == "" || !hasRecord(game.RecordKeys, root.PassivePath) {
		problems = append(problems, "ゲーム内パッシブレコードがありません")
	}

	for i, path := range root.SpellPaths {
		if i >= spellCount {
			break
		}

		if path == "" || !hasRecord(game.RecordKeys, path) {
			problems = append(problems, fmt.Sprintf("ゲーム内スキルレコードがありません: %d", i+1))
		}
	}

	return problems
}

func auditChampion(ctx context.Context, src Source, patch string, id int, summaries map[int]champion.Summary) []string {
	summary, exists := summaries[id]
	if !exists || summary.Name == "" || summary.Alias == "" {
		return []string{"CommunityDragon の日本語一覧にありません"}
	}

	detail, err := src.Detail(ctx, patch, id)
	if err != nil {
		return []string{err.Error()}
	}

	if detail.ID != id {
		return []string{"詳細データの ID が一致しません"}
	}

	game, err := src.GameData(ctx, patch, summary.Alias)
	if err != nil {
		return []string{err.Error()}
	}

	return append(checkDetail(detail), checkGameData(game, summary.Alias)...)
}

func Run(ctx context.Context, src Source, version string) (Report, error) {
	if version == "" {
		var err error

		version, err = src.LatestVersion(ctx)
		if err != nil {
			return Report{}, err
		}
	}

	patch, err := champion.PatchFromVersion(version)
	if err != nil {
		return Report{}, err
	}

	roster, err := src.Roster(ctx, version)
	if err != nil {
		return Report{}, err
	}

	summaries, err := src.Summaries(ctx, patch)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		DataDragonVersion: version,
		Patch:             patch,
		ChampionCount:     len(roster),
		StructuralPasses:  0,
		Failures:          []Failure{},
		Note:              "構造検証のみ。数値・係数・変身後スキルの正確さは未検証です。",
	}
	for _, id := range slices.Sorted(maps.Keys(roster)) {
		problems := auditChampion(ctx, src, patch, id, summaries)
		if len(problems) == 0 {
			report.StructuralPasses++
			continue
		}

		for _, problem := range problems {
			report.Failures = append(report.Failures, Failure{roster[id].ID, problem})
		}
	}

	return report, nil
}
