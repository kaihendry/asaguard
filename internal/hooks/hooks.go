package hooks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HookEntry matches the Claude Code settings.json hooks schema.
type HookEntry struct {
	Matcher string   `json:"matcher,omitempty"`
	Hooks   []Hook   `json:"hooks"`
}

type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type settings struct {
	Hooks map[string][]HookEntry `json:"hooks,omitempty"`
	// Preserve all other fields
	Extra map[string]json.RawMessage `json:"-"`
}

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func readRaw() (map[string]json.RawMessage, error) {
	data, err := os.ReadFile(settingsPath())
	if os.IsNotExist(err) {
		return map[string]json.RawMessage{}, nil
	}
	if err != nil {
		return nil, err
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func writeAtomic(m map[string]json.RawMessage) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	path := settingsPath()
	tmp := path + ".asaguard.tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func getHooks(m map[string]json.RawMessage, event string) ([]HookEntry, error) {
	hooksRaw, ok := m["hooks"]
	if !ok {
		return nil, nil
	}
	var allHooks map[string][]HookEntry
	if err := json.Unmarshal(hooksRaw, &allHooks); err != nil {
		return nil, err
	}
	return allHooks[event], nil
}

func setHooks(m map[string]json.RawMessage, event string, entries []HookEntry) error {
	var allHooks map[string][]HookEntry
	if raw, ok := m["hooks"]; ok {
		json.Unmarshal(raw, &allHooks)
	}
	if allHooks == nil {
		allHooks = map[string][]HookEntry{}
	}
	if len(entries) == 0 {
		delete(allHooks, event)
	} else {
		allHooks[event] = entries
	}
	raw, err := json.Marshal(allHooks)
	if err != nil {
		return err
	}
	m["hooks"] = raw
	return nil
}

// Confirm shows a diff and asks the user for y/n confirmation.
func Confirm(description string) bool {
	fmt.Println(description)
	fmt.Print("Proceed? [y/N] ")
	sc := bufio.NewScanner(os.Stdin)
	if sc.Scan() {
		return strings.ToLower(strings.TrimSpace(sc.Text())) == "y"
	}
	return false
}

// Install adds a PreToolUse hook entry with the given matcher and command.
// It checks for duplicates and skips if already present.
func Install(event, matcher, command, description string) error {
	m, err := readRaw()
	if err != nil {
		return err
	}

	entries, err := getHooks(m, event)
	if err != nil {
		return err
	}

	// Idempotency check
	for _, e := range entries {
		for _, h := range e.Hooks {
			if h.Command == command {
				fmt.Printf("%s: already installed\n", description)
				return nil
			}
		}
	}

	newEntry := HookEntry{
		Matcher: matcher,
		Hooks:   []Hook{{Type: "command", Command: command}},
	}

	proposal := fmt.Sprintf("Add %s hook:\n  event: %s\n  matcher: %s\n  command: %s", description, event, matcher, command)
	if !Confirm(proposal) {
		fmt.Println("Aborted.")
		return nil
	}

	entries = append(entries, newEntry)
	if err := setHooks(m, event, entries); err != nil {
		return err
	}
	if err := writeAtomic(m); err != nil {
		return err
	}
	fmt.Printf("Installed %s hook.\n", description)
	return nil
}

// Uninstall removes all hook entries whose command matches the given command string.
func Uninstall(event, command, description string) error {
	m, err := readRaw()
	if err != nil {
		return err
	}

	entries, err := getHooks(m, event)
	if err != nil {
		return err
	}

	var kept []HookEntry
	removed := 0
	for _, e := range entries {
		var keptHooks []Hook
		for _, h := range e.Hooks {
			if h.Command == command {
				removed++
			} else {
				keptHooks = append(keptHooks, h)
			}
		}
		if len(keptHooks) > 0 {
			e.Hooks = keptHooks
			kept = append(kept, e)
		}
	}

	if removed == 0 {
		fmt.Printf("no %s hook found\n", description)
		return nil
	}

	if err := setHooks(m, event, kept); err != nil {
		return err
	}
	if err := writeAtomic(m); err != nil {
		return err
	}
	fmt.Printf("Removed %s hook.\n", description)
	return nil
}
