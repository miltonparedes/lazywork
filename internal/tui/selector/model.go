package selector

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/miltonparedes/lazywork/internal/git"
)

// ActionResult represents the action selected by the user.
type ActionResult struct {
	Type string // "go", "delete", "add", "use", "finish", "quit"
	Path string
	Name string
}

// Model is the Bubbletea model for the worktree selector.
type Model struct {
	worktrees   []git.Worktree
	cursor      int
	currentPath string
	action      *ActionResult
	quitting    bool
}

func New(worktrees []git.Worktree, currentPath string) Model {
	var filtered []git.Worktree
	for _, wt := range worktrees {
		if !wt.Bare && strings.Contains(wt.Path, string(filepath.Separator)+".worktrees"+string(filepath.Separator)) {
			filtered = append(filtered, wt)
		}
	}

	return Model{
		worktrees:   filtered,
		cursor:      0,
		currentPath: currentPath,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			m.action = &ActionResult{Type: "quit"}
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.worktrees)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.worktrees) > 0 {
				wt := m.worktrees[m.cursor]
				m.action = &ActionResult{
					Type: "go",
					Path: wt.Path,
					Name: filepath.Base(wt.Path),
				}
				m.quitting = true
				return m, tea.Quit
			}

		case "d":
			if len(m.worktrees) > 0 {
				wt := m.worktrees[m.cursor]
				m.action = &ActionResult{
					Type: "delete",
					Path: wt.Path,
					Name: filepath.Base(wt.Path),
				}
				m.quitting = true
				return m, tea.Quit
			}

		case "a":
			m.action = &ActionResult{Type: "add"}
			m.quitting = true
			return m, tea.Quit

		case "u":
			if len(m.worktrees) > 0 {
				wt := m.worktrees[m.cursor]
				m.action = &ActionResult{
					Type: "use",
					Path: wt.Path,
					Name: filepath.Base(wt.Path),
				}
				m.quitting = true
				return m, tea.Quit
			}

		case "f":
			if len(m.worktrees) > 0 {
				wt := m.worktrees[m.cursor]
				m.action = &ActionResult{
					Type: "finish",
					Path: wt.Path,
					Name: filepath.Base(wt.Path),
				}
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("Worktrees (%d):", len(m.worktrees))))
	b.WriteString("\n")

	if len(m.worktrees) == 0 {
		b.WriteString(dimStyle.Render("  No worktrees found. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		for i, wt := range m.worktrees {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("> ")
			}

			name := filepath.Base(wt.Path)
			branch := wt.Branch
			if branch == "" && len(wt.Head) >= 7 {
				branch = fmt.Sprintf("detached:%s", wt.Head[:7])
			}

			var line string
			if i == m.cursor {
				line = selectedStyle.Render(fmt.Sprintf("%s (%s)", name, branch))
			} else {
				line = normalStyle.Render(fmt.Sprintf("%s (%s)", name, branch))
			}

			isCurrent := strings.HasPrefix(m.currentPath, wt.Path)

			b.WriteString(cursor)
			b.WriteString(line)
			if isCurrent {
				b.WriteString(currentMarker.Render(""))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/↓ navigate  enter go  d delete  a add  u use  f finish  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) Action() *ActionResult {
	return m.action
}

func (m Model) SelectedWorktree() *git.Worktree {
	if len(m.worktrees) == 0 || m.cursor >= len(m.worktrees) {
		return nil
	}
	return &m.worktrees[m.cursor]
}
