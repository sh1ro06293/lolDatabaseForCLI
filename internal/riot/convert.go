package riot

import (
	"encoding/json"
	"strings"

	"github.com/sh1ro06293/lolDatabaseForCLI/internal/champion"
)

func toDetail(raw detail) champion.Detail {
	spells := make([]champion.Spell, 0, len(raw.Spells))
	for _, s := range raw.Spells {
		spells = append(spells, champion.Spell{
			Key:                s.SpellKey,
			Name:               s.Name,
			Description:        s.Description,
			DynamicDescription: s.DynamicDescription,
		})
	}

	return champion.Detail{
		ID:      raw.ID,
		Passive: champion.Passive{Name: raw.Passive.Name, Description: raw.Passive.Description},
		Spells:  spells,
	}
}

func findRecord(records map[string]json.RawMessage, key string) (json.RawMessage, bool) {
	for recordKey, raw := range records {
		if strings.EqualFold(recordKey, key) {
			return raw, true
		}
	}

	return nil, false
}
