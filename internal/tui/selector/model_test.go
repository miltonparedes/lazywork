package selector

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/miltonparedes/lazywork/internal/git"
)

func makeWorktrees() []git.Worktree {
	return []git.Worktree{
		{Path: "/repo/.worktrees/feature-a", Branch: "feature-a", Head: "abc1234"},
		{Path: "/repo/.worktrees/feature-b", Branch: "feature-b", Head: "def5678"},
		{Path: "/repo/.worktrees/bugfix-x", Branch: "bugfix-x", Head: "ghi9012"},
	}
}

func TestNew(t *testing.T) {
	worktrees := makeWorktrees()
	model := New(worktrees, "/repo/.worktrees/feature-a")

	if len(model.worktrees) != 3 {
		t.Errorf("expected 3 worktrees, got %d", len(model.worktrees))
	}

	if model.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", model.cursor)
	}

	if model.currentPath != "/repo/.worktrees/feature-a" {
		t.Errorf("expected currentPath to be set")
	}

	if model.action != nil {
		t.Error("expected action to be nil initially")
	}

	if model.quitting {
		t.Error("expected quitting to be false initially")
	}
}

func TestNewFiltersBareRepos(t *testing.T) {
	worktrees := []git.Worktree{
		{Path: "/repo", Bare: true},
		{Path: "/repo/.worktrees/feature-a", Branch: "feature-a"},
	}

	model := New(worktrees, "/repo")

	if len(model.worktrees) != 1 {
		t.Errorf("expected 1 worktree (bare filtered), got %d", len(model.worktrees))
	}
}

func TestUpdateNavigationDown(t *testing.T) {
	model := New(makeWorktrees(), "")

	// Press down
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := newModel.(Model)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.cursor)
	}

	// Press j (vim down)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)

	if m.cursor != 2 {
		t.Errorf("expected cursor at 2 after j, got %d", m.cursor)
	}

	// Press down at bottom - should stay
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)

	if m.cursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", m.cursor)
	}
}

func TestUpdateNavigationUp(t *testing.T) {
	model := New(makeWorktrees(), "")
	model.cursor = 2

	// Press up
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m := newModel.(Model)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after up, got %d", m.cursor)
	}

	// Press k (vim up)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after k, got %d", m.cursor)
	}

	// Press up at top - should stay
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)

	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", m.cursor)
	}
}

func TestUpdateEnterAction(t *testing.T) {
	model := New(makeWorktrees(), "")
	model.cursor = 1

	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.action == nil {
		t.Fatal("expected action to be set")
	}

	if m.action.Type != "go" {
		t.Errorf("expected action type 'go', got %q", m.action.Type)
	}

	if m.action.Path != "/repo/.worktrees/feature-b" {
		t.Errorf("expected path feature-b, got %q", m.action.Path)
	}

	if m.action.Name != "feature-b" {
		t.Errorf("expected name feature-b, got %q", m.action.Name)
	}

	if !m.quitting {
		t.Error("expected quitting to be true")
	}

	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestUpdateDeleteAction(t *testing.T) {
	model := New(makeWorktrees(), "")

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m := newModel.(Model)

	if m.action == nil {
		t.Fatal("expected action to be set")
	}

	if m.action.Type != "delete" {
		t.Errorf("expected action type 'delete', got %q", m.action.Type)
	}
}

func TestUpdateAddAction(t *testing.T) {
	model := New(makeWorktrees(), "")

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m := newModel.(Model)

	if m.action == nil {
		t.Fatal("expected action to be set")
	}

	if m.action.Type != "add" {
		t.Errorf("expected action type 'add', got %q", m.action.Type)
	}
}

func TestUpdateUseAction(t *testing.T) {
	model := New(makeWorktrees(), "")

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m := newModel.(Model)

	if m.action == nil {
		t.Fatal("expected action to be set")
	}

	if m.action.Type != "use" {
		t.Errorf("expected action type 'use', got %q", m.action.Type)
	}
}

func TestUpdateFinishAction(t *testing.T) {
	model := New(makeWorktrees(), "")

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m := newModel.(Model)

	if m.action == nil {
		t.Fatal("expected action to be set")
	}

	if m.action.Type != "finish" {
		t.Errorf("expected action type 'finish', got %q", m.action.Type)
	}
}

func TestUpdateQuitActions(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"esc key", tea.KeyMsg{Type: tea.KeyEscape}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New(makeWorktrees(), "")

			newModel, _ := model.Update(tt.key)
			m := newModel.(Model)

			if m.action == nil {
				t.Fatal("expected action to be set")
			}

			if m.action.Type != "quit" {
				t.Errorf("expected action type 'quit', got %q", m.action.Type)
			}

			if !m.quitting {
				t.Error("expected quitting to be true")
			}
		})
	}
}

func TestUpdateEmptyWorktrees(t *testing.T) {
	model := New([]git.Worktree{}, "")

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(Model)

	if m.action != nil {
		t.Error("expected no action on empty list")
	}

	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = newModel.(Model)

	if m.action != nil {
		t.Error("expected no action on empty list")
	}
}

func TestAction(t *testing.T) {
	model := New(makeWorktrees(), "")

	if model.Action() != nil {
		t.Error("expected nil action initially")
	}

	model.action = &ActionResult{Type: "go", Path: "/path"}

	action := model.Action()
	if action == nil || action.Type != "go" {
		t.Error("expected action to be returned")
	}
}

func TestSelectedWorktree(t *testing.T) {
	model := New(makeWorktrees(), "")

	wt := model.SelectedWorktree()
	if wt == nil {
		t.Fatal("expected worktree")
	}

	if wt.Branch != "feature-a" {
		t.Errorf("expected feature-a, got %s", wt.Branch)
	}

	model.cursor = 2
	wt = model.SelectedWorktree()
	if wt.Branch != "bugfix-x" {
		t.Errorf("expected bugfix-x, got %s", wt.Branch)
	}
}

func TestSelectedWorktreeEmpty(t *testing.T) {
	model := New([]git.Worktree{}, "")

	wt := model.SelectedWorktree()
	if wt != nil {
		t.Error("expected nil for empty worktrees")
	}
}

func TestViewNotQuitting(t *testing.T) {
	model := New(makeWorktrees(), "/repo/.worktrees/feature-a")

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	if !contains(view, "Worktrees") {
		t.Error("expected view to contain 'Worktrees'")
	}

	if !contains(view, "feature-a") {
		t.Error("expected view to contain 'feature-a'")
	}

	if !contains(view, "navigate") {
		t.Error("expected view to contain help text")
	}

	if !contains(view, "current") {
		t.Error("expected view to mark current worktree")
	}
}

func TestViewQuitting(t *testing.T) {
	model := New(makeWorktrees(), "")
	model.quitting = true

	view := model.View()

	if view != "" {
		t.Errorf("expected empty view when quitting, got %q", view)
	}
}

func TestViewEmpty(t *testing.T) {
	model := New([]git.Worktree{}, "")

	view := model.View()

	if !contains(view, "No worktrees found") {
		t.Error("expected empty state message")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
