package audit

import (
	"context"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
)

// Source は監査に使うデータの取得元
type Source interface {
	LatestVersion(ctx context.Context) (string, error)
	Roster(ctx context.Context, version string) (map[int]champion.RiotEntry, error)
	Summaries(ctx context.Context, patch string) (map[int]champion.Summary, error)
	Detail(ctx context.Context, patch string, id int) (champion.Detail, error)
	GameData(ctx context.Context, patch, alias string) (champion.GameData, error)
}

type Failure struct {
	Champion string `json:"champion"`
	Reason   string `json:"reason"`
}

type Report struct {
	DataDragonVersion string    `json:"dataDragonVersion"`
	Patch             string    `json:"patch"`
	ChampionCount     int       `json:"championCount"`
	StructuralPasses  int       `json:"structuralPasses"`
	Failures          []Failure `json:"failures"`
	Note              string    `json:"note"`
}
