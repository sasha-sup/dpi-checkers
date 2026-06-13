package cli

import (
	"encoding/json"
	"os"
)

// stateFile holds the persisted notification state between runs.
// It lives next to the binary (the process chdirs there in release builds).
const stateFile = "dpich-state.json"

// state is what we remember across runs to make notifications stateful:
// which target keys were down last time, and the date of the last heartbeat.
type state struct {
	Down          []string `json:"down"`           // keys (section/target) currently considered down
	HeartbeatDate string   `json:"heartbeat_date"` // UTC date (YYYY-MM-DD) of last heartbeat sent
}

func loadState() state {
	var s state
	b, err := os.ReadFile(stateFile)
	if err != nil {
		return s // first run / unreadable -> empty
	}
	_ = json.Unmarshal(b, &s)
	return s
}

func saveState(s state) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(stateFile, b, 0600)
}
