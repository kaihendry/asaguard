package transcripts

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Entry represents one line from a Claude Code JSONL transcript.
type Entry struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	SessionID string    `json:"sessionId"`

	// Tool call fields
	ToolName  string          `json:"toolName"`
	ToolInput json.RawMessage `json:"toolInput"`

	// Usage fields (nested under message.usage for assistant turns)
	Message *struct {
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			CacheRead    int `json:"cache_read_input_tokens"`
			CacheWrite   int `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`

	// Session metadata
	CLIArgs []string `json:"cliArgs"`
}

// Session holds aggregated data for one session file.
type Session struct {
	ID      string
	Path    string
	Entries []Entry
}

// Walk visits all JSONL transcript files under ~/.claude/projects/.
// If since is non-zero, files older than since are skipped.
func Walk(since time.Time, fn func(*Session) error) error {
	home, _ := os.UserHomeDir()
	root := filepath.Join(home, ".claude", "projects")

	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return err
		}
		if !since.IsZero() {
			info, err := d.Info()
			if err == nil && info.ModTime().Before(since) {
				return nil
			}
		}
		sess, err := parseFile(path)
		if err != nil {
			return nil // skip unparseable files, don't abort walk
		}
		return fn(sess)
	})
}

func parseFile(path string) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sess := &Session{Path: path}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20) // 1 MiB per line
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue // tolerate unknown fields
		}
		if sess.ID == "" && e.SessionID != "" {
			sess.ID = e.SessionID
		}
		sess.Entries = append(sess.Entries, e)
	}
	if sess.ID == "" {
		sess.ID = filepath.Base(path)
	}
	return sess, sc.Err()
}
