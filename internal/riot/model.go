package riot

type rosterResponse struct {
	Data map[string]rosterEntry `json:"data"`
}

type rosterEntry struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

type summary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type passive struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type spell struct {
	SpellKey           string `json:"spellKey"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	DynamicDescription string `json:"dynamicDescription"`
}

type detail struct {
	ID      int     `json:"id"`
	Passive passive `json:"passive"`
	Spells  []spell `json:"spells"`
}

type gameRoot struct {
	Spells  []string `json:"spells"`
	Passive string   `json:"mCharacterPassiveSpell"`
}
