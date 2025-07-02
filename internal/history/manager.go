package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8x/internal/config"
)

// Entry represents a k8x session entry
type Entry struct {
	ID        string    `json:"id"`
	Goal      string    `json:"goal"`
	Steps     []Step    `json:"steps"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "success", "error", "pending"
}

// Step represents a single step in a k8x session
type Step struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	Output      string `json:"output,omitempty"`
	UndoCommand string `json:"undo_command,omitempty"`
	Type        string `json:"type"` // "step", "exploratory", "question"
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

// Save saves a history entry to disk in .k8x format
func (m *Manager) Save(entry *Entry) error {
	if entry.ID == "" {
		entry.ID = generateID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Create kebab-case filename
	goalKebab := strings.ToLower(strings.ReplaceAll(entry.Goal, " ", "-"))
	goalKebab = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(goalKebab, "")
	filename := fmt.Sprintf("%s-%s.k8x", goalKebab, entry.Timestamp.Format("20060102-150405"))

	filepath := filepath.Join(m.historyDir, filename)

	var content strings.Builder

	// Write shebang and goal
	content.WriteString("#!/bin/k8x\n")
	content.WriteString(fmt.Sprintf("#$ %s\n\n", entry.Goal))

	// Write steps
	for i, step := range entry.Steps {
		switch step.Type {
		case "exploratory":
			content.WriteString(fmt.Sprintf("#~ %s\n", step.Description))
		case "question":
			content.WriteString(fmt.Sprintf("#? %s\n", step.Description))
		default:
			content.WriteString(fmt.Sprintf("# %d. %s\n", i+1, step.Description))
		}

		if step.Command != "" {
			content.WriteString(step.Command + "\n")
		}

		if step.Output != "" {
			outputLines := strings.Split(step.Output, "\n")
			for _, line := range outputLines {
				if line != "" {
					content.WriteString(fmt.Sprintf("#> %s\n", line))
				}
			}
		}

		if step.UndoCommand != "" {
			content.WriteString(fmt.Sprintf("#- %s\n", step.UndoCommand))
		}

		content.WriteString("\n")
	}

	if err := os.WriteFile(filepath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// Load loads a history entry by filename from .k8x format
func (m *Manager) Load(filename string) (*Entry, error) {
	filepath := filepath.Join(m.historyDir, filename)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", closeErr)
		}
	}()

	entry := &Entry{
		ID:    generateID(),
		Steps: []Step{},
	}

	scanner := bufio.NewScanner(file)
	var currentStep *Step

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#$ ") {
			entry.Goal = strings.TrimPrefix(line, "#$ ")
		} else if strings.HasPrefix(line, "# ") {
			// New step
			if currentStep != nil {
				entry.Steps = append(entry.Steps, *currentStep)
			}
			currentStep = &Step{
				Description: strings.TrimPrefix(line, "# "),
				Type:        "step",
			}
		} else if strings.HasPrefix(line, "#~ ") {
			// Exploratory step
			if currentStep != nil {
				entry.Steps = append(entry.Steps, *currentStep)
			}
			currentStep = &Step{
				Description: strings.TrimPrefix(line, "#~ "),
				Type:        "exploratory",
			}
		} else if strings.HasPrefix(line, "#? ") {
			// Question step
			if currentStep != nil {
				entry.Steps = append(entry.Steps, *currentStep)
			}
			currentStep = &Step{
				Description: strings.TrimPrefix(line, "#? "),
				Type:        "question",
			}
		} else if strings.HasPrefix(line, "#> ") {
			// Output
			if currentStep != nil {
				output := strings.TrimPrefix(line, "#> ")
				if currentStep.Output == "" {
					currentStep.Output = output
				} else {
					currentStep.Output += "\n" + output
				}
			}
		} else if strings.HasPrefix(line, "#- ") {
			// Undo command
			if currentStep != nil {
				currentStep.UndoCommand = strings.TrimPrefix(line, "#- ")
			}
		} else if line != "" && !strings.HasPrefix(line, "#") {
			// Command
			if currentStep != nil {
				currentStep.Command = line
			}
		}
	}

	// Add the last step
	if currentStep != nil {
		entry.Steps = append(entry.Steps, *currentStep)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan history file: %w", err)
	}

	return entry, nil
}

// List returns a list of history files
func (m *Manager) List() ([]string, error) {
	files, err := os.ReadDir(m.historyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var historyFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".k8x" {
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

// AddStep adds a new step to an existing entry and saves it
func (m *Manager) AddStep(entry *Entry, step Step) error {
	entry.Steps = append(entry.Steps, step)
	return m.Save(entry)
}

// UpdateEntry updates an existing entry by re-saving it
func (m *Manager) UpdateEntry(entry *Entry) error {
	return m.Save(entry)
}

// generateID generates a unique ID for a history entry
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
