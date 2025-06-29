package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shankgan/k8x/internal/config"
)

// Entry represents a command history entry
type Entry struct {
	ID        string         `json:"id"`
	Goal      string         `json:"goal"`
	Command   string         `json:"command"`
	Args      []string       `json:"args"`
	Timestamp time.Time      `json:"timestamp"`
	Status    string         `json:"status"` // "success", "error", "pending"
	Output    string         `json:"output,omitempty"`
	Error     string         `json:"error,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// Manager handles command history operations
type Manager struct {
	historyDir string
}

// NewManager creates a new history manager
func NewManager() (*Manager, error) {
	historyDir, err := config.GetHistoryDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get history directory: %w", err)
	}

	// Ensure history directory exists
	if err := os.MkdirAll(historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return &Manager{
		historyDir: historyDir,
	}, nil
}

// Save saves a history entry to disk
func (m *Manager) Save(entry *Entry) error {
	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	filename := fmt.Sprintf("%s-%s.json", entry.Goal, entry.Timestamp.Format("20060102-150405"))
	filename = sanitizeFilename(filename)

	filepath := filepath.Join(m.historyDir, filename)

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history entry: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// Load loads a history entry by filename
func (m *Manager) Load(filename string) (*Entry, error) {
	filepath := filepath.Join(m.historyDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history entry: %w", err)
	}

	return &entry, nil
}

// List returns a list of history files
func (m *Manager) List() ([]string, error) {
	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var historyFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			historyFiles = append(historyFiles, file.Name())
		}
	}

	return historyFiles, nil
}

// Delete removes a history file
func (m *Manager) Delete(filename string) error {
	filepath := filepath.Join(m.historyDir, filename)
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete history file: %w", err)
	}
	return nil
}

// generateID generates a unique ID for a history entry
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(filename string) string {
	// Replace invalid characters with hyphens
	invalid := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|', ' '}
	runes := []rune(filename)

	for i, r := range runes {
		for _, inv := range invalid {
			if r == inv {
				runes[i] = '-'
				break
			}
		}
	}

	return string(runes)
}
