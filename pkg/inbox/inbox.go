package inbox

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Item struct {
	ID      string    `json:"id"`
	Text    string    `json:"text"`
	Created time.Time `json:"created"`
}

type Store struct {
	path  string
	items []Item
}

func NewStoreForVault(vaultPath string) *Store {
	return &Store{path: PathForVault(vaultPath)}
}

func NewStoreWithPath(path string) *Store {
	return &Store{path: path}
}

// PathForVault returns the inbox file path for a given vault, stored under the
// lazyruin config directory keyed by a hash of the vault path.
func PathForVault(vaultPath string) string {
	hash := sha256.Sum256([]byte(vaultPath))
	name := fmt.Sprintf("%x.json", hash[:8])
	return filepath.Join(configDir(), "inboxes", name)
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.items = nil
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.items)
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) Add(text string) {
	s.items = append(s.items, Item{
		ID:      randomHex(6),
		Text:    text,
		Created: time.Now(),
	})
}

func (s *Store) Delete(id string) {
	for i, item := range s.items {
		if item.ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

func (s *Store) Items() []Item {
	sorted := make([]Item, len(s.items))
	copy(sorted, s.items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Created.After(sorted[j].Created)
	})
	return sorted
}

func (s *Store) Len() int {
	return len(s.items)
}

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "lazyruin")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazyruin")
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
