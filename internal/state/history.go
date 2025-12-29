package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// WorktreeHistory tracks navigation history across repositories.
type WorktreeHistory struct {
	RepoHistory map[string]RepoState `json:"repo_history"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// RepoState tracks the current and previous worktree for a repository.
type RepoState struct {
	Previous string `json:"previous"`
	Current  string `json:"current"`
}

// HistoryPath returns the path to the history file.
func HistoryPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "lazywork", "history.json")
}

// LoadHistory loads the worktree history from disk.
// Returns an empty history if the file doesn't exist.
func LoadHistory() (*WorktreeHistory, error) {
	historyPath := HistoryPath()

	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return &WorktreeHistory{
			RepoHistory: make(map[string]RepoState),
		}, nil
	}

	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, err
	}

	var history WorktreeHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	if history.RepoHistory == nil {
		history.RepoHistory = make(map[string]RepoState)
	}

	return &history, nil
}

// Save persists the worktree history to disk.
func (h *WorktreeHistory) Save() error {
	historyPath := HistoryPath()

	historyDir := filepath.Dir(historyPath)
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		return err
	}

	h.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath, data, 0o644)
}

// RecordVisit records a visit to a worktree path for the given repository.
// The current path becomes the previous, and the new path becomes current.
func (h *WorktreeHistory) RecordVisit(repoRoot, worktreePath string) {
	if h.RepoHistory == nil {
		h.RepoHistory = make(map[string]RepoState)
	}

	state := h.RepoHistory[repoRoot]

	// Only update if we're actually changing worktrees
	if state.Current != worktreePath {
		state.Previous = state.Current
		state.Current = worktreePath
		h.RepoHistory[repoRoot] = state
	}
}

// GetPrevious returns the previous worktree path for the given repository.
// Returns an empty string if there's no previous worktree.
func (h *WorktreeHistory) GetPrevious(repoRoot string) string {
	if h.RepoHistory == nil {
		return ""
	}
	return h.RepoHistory[repoRoot].Previous
}

// GetCurrent returns the current worktree path for the given repository.
// Returns an empty string if there's no current worktree recorded.
func (h *WorktreeHistory) GetCurrent(repoRoot string) string {
	if h.RepoHistory == nil {
		return ""
	}
	return h.RepoHistory[repoRoot].Current
}
