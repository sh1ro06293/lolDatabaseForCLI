package champion

// RiotEntry は Data Dragon のチャンピオン一覧に含まれる識別子
type RiotEntry struct {
	ID string
}

// Summary は CommunityDragon の日本語チャンピオン一覧の項目
type Summary struct {
	ID    int
	Name  string
	Alias string
}

type Passive struct {
	Name        string
	Description string
}

type Spell struct {
	Key                string
	Name               string
	Description        string
	DynamicDescription string
}

// Detail は CommunityDragon の日本語スキル情報
type Detail struct {
	ID      int
	Passive Passive
	Spells  []Spell
}

// GameRoot はゲーム内データで通常形態のスキルを参照するレコード
type GameRoot struct {
	SpellPaths  []string
	PassivePath string
}

// GameData はゲーム内データのうち、構造の確認に使う部分
// Root が nil のときは、通常形態のレコードが無い
type GameData struct {
	Root       *GameRoot
	RecordKeys map[string]struct{}
}
