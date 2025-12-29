package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHistoryPath(t *testing.T) {
	path := HistoryPath()
	if path == "" {
		t.Error("HistoryPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("HistoryPath() returned non-absolute path: %s", path)
	}
	if filepath.Base(path) != "history.json" {
		t.Errorf("HistoryPath() should end with history.json, got: %s", path)
	}
}

func TestLoadHistoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	history, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error: %v", err)
	}

	if history == nil {
		t.Fatal("LoadHistory() returned nil")
	}

	if history.RepoHistory == nil {
		t.Error("LoadHistory() returned nil RepoHistory map")
	}

	if len(history.RepoHistory) != 0 {
		t.Errorf("LoadHistory() returned non-empty RepoHistory: %v", history.RepoHistory)
	}
}

func TestHistorySaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	history := &WorktreeHistory{
		RepoHistory: map[string]RepoState{
			"/path/to/repo": {
				Previous: "/path/to/repo/.worktrees/feature-a",
				Current:  "/path/to/repo/.worktrees/feature-b",
			},
		},
	}

	if err := history.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	historyPath := HistoryPath()
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		t.Fatalf("History file was not created at %s", historyPath)
	}

	loaded, err := LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory() error: %v", err)
	}

	state := loaded.RepoHistory["/path/to/repo"]
	if state.Previous != "/path/to/repo/.worktrees/feature-a" {
		t.Errorf("Previous = %q, want %q", state.Previous, "/path/to/repo/.worktrees/feature-a")
	}
	if state.Current != "/path/to/repo/.worktrees/feature-b" {
		t.Errorf("Current = %q, want %q", state.Current, "/path/to/repo/.worktrees/feature-b")
	}
}

func TestRecordVisit(t *testing.T) {
	history := &WorktreeHistory{}

	repoRoot := "/path/to/repo"
	worktree1 := "/path/to/repo/.worktrees/feature-a"
	worktree2 := "/path/to/repo/.worktrees/feature-b"
	worktree3 := "/path/to/repo/.worktrees/feature-c"

	history.RecordVisit(repoRoot, worktree1)
	if history.GetCurrent(repoRoot) != worktree1 {
		t.Errorf("Current = %q, want %q", history.GetCurrent(repoRoot), worktree1)
	}
	if history.GetPrevious(repoRoot) != "" {
		t.Errorf("Previous = %q, want empty", history.GetPrevious(repoRoot))
	}

	history.RecordVisit(repoRoot, worktree2)
	if history.GetCurrent(repoRoot) != worktree2 {
		t.Errorf("Current = %q, want %q", history.GetCurrent(repoRoot), worktree2)
	}
	if history.GetPrevious(repoRoot) != worktree1 {
		t.Errorf("Previous = %q, want %q", history.GetPrevious(repoRoot), worktree1)
	}

	history.RecordVisit(repoRoot, worktree3)
	if history.GetCurrent(repoRoot) != worktree3 {
		t.Errorf("Current = %q, want %q", history.GetCurrent(repoRoot), worktree3)
	}
	if history.GetPrevious(repoRoot) != worktree2 {
		t.Errorf("Previous = %q, want %q", history.GetPrevious(repoRoot), worktree2)
	}
}

func TestRecordVisitSameWorktree(t *testing.T) {
	history := &WorktreeHistory{}

	repoRoot := "/path/to/repo"
	worktree1 := "/path/to/repo/.worktrees/feature-a"
	worktree2 := "/path/to/repo/.worktrees/feature-b"

	history.RecordVisit(repoRoot, worktree1)
	history.RecordVisit(repoRoot, worktree2)

	history.RecordVisit(repoRoot, worktree2)
	if history.GetCurrent(repoRoot) != worktree2 {
		t.Errorf("Current = %q, want %q", history.GetCurrent(repoRoot), worktree2)
	}
	if history.GetPrevious(repoRoot) != worktree1 {
		t.Errorf("Previous = %q, want %q (should not change)", history.GetPrevious(repoRoot), worktree1)
	}
}

func TestMultipleRepos(t *testing.T) {
	history := &WorktreeHistory{}

	repo1 := "/path/to/repo1"
	repo2 := "/path/to/repo2"
	worktree1a := "/path/to/repo1/.worktrees/feature-a"
	worktree1b := "/path/to/repo1/.worktrees/feature-b"
	worktree2a := "/path/to/repo2/.worktrees/feature-x"
	worktree2b := "/path/to/repo2/.worktrees/feature-y"

	history.RecordVisit(repo1, worktree1a)
	history.RecordVisit(repo1, worktree1b)

	history.RecordVisit(repo2, worktree2a)
	history.RecordVisit(repo2, worktree2b)

	if history.GetCurrent(repo1) != worktree1b {
		t.Errorf("repo1 Current = %q, want %q", history.GetCurrent(repo1), worktree1b)
	}
	if history.GetPrevious(repo1) != worktree1a {
		t.Errorf("repo1 Previous = %q, want %q", history.GetPrevious(repo1), worktree1a)
	}

	if history.GetCurrent(repo2) != worktree2b {
		t.Errorf("repo2 Current = %q, want %q", history.GetCurrent(repo2), worktree2b)
	}
	if history.GetPrevious(repo2) != worktree2a {
		t.Errorf("repo2 Previous = %q, want %q", history.GetPrevious(repo2), worktree2a)
	}
}

func TestGetPreviousUnknownRepo(t *testing.T) {
	history := &WorktreeHistory{
		RepoHistory: make(map[string]RepoState),
	}

	previous := history.GetPrevious("/unknown/repo")
	if previous != "" {
		t.Errorf("GetPrevious() for unknown repo = %q, want empty", previous)
	}
}

func TestGetCurrentUnknownRepo(t *testing.T) {
	history := &WorktreeHistory{
		RepoHistory: make(map[string]RepoState),
	}

	current := history.GetCurrent("/unknown/repo")
	if current != "" {
		t.Errorf("GetCurrent() for unknown repo = %q, want empty", current)
	}
}

func TestNilRepoHistory(t *testing.T) {
	history := &WorktreeHistory{}

	// Should not panic
	previous := history.GetPrevious("/some/repo")
	if previous != "" {
		t.Errorf("GetPrevious() with nil RepoHistory = %q, want empty", previous)
	}

	current := history.GetCurrent("/some/repo")
	if current != "" {
		t.Errorf("GetCurrent() with nil RepoHistory = %q, want empty", current)
	}

	// RecordVisit should initialize the map
	history.RecordVisit("/some/repo", "/some/worktree")
	if history.RepoHistory == nil {
		t.Error("RecordVisit() did not initialize RepoHistory")
	}
}
