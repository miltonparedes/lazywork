package tui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/miltonparedes/lazywork/internal/git"
)

// ANSI base colors (0-7) adapt to user's terminal theme
var (
	ColorPrimary  = lipgloss.Color("4") // blue
	ColorSelected = lipgloss.Color("2") // green
	ColorNormal   = lipgloss.Color("7") // foreground
	ColorDim      = lipgloss.Color("8") // bright black
	ColorAccent   = lipgloss.Color("3") // yellow
)

func Theme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Title = t.Focused.Title.Foreground(ColorPrimary).Bold(true)
	t.Blurred.Title = t.Blurred.Title.Foreground(ColorDim)

	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(ColorSelected).Bold(true)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(ColorNormal)

	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(ColorSelected).SetString("> ")

	t.Help.ShortKey = t.Help.ShortKey.Foreground(ColorDim)
	t.Help.ShortDesc = t.Help.ShortDesc.Foreground(ColorDim)

	return t
}

func BranchNameForm(name *string) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Branch name").
				Placeholder("feature-xyz").
				Value(name),
		),
	).WithTheme(Theme())
}

func ConfirmForm(message string, confirmed *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Value(confirmed),
		),
	).WithTheme(Theme())
}

func SelectForm(title string, options []string, selected *string) *huh.Form {
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(opts...).
				Value(selected),
		),
	).WithTheme(Theme())
}

// Returns the selected worktree name (basename of path)
func WorktreeSelectForm(worktrees []git.Worktree, selected *string) *huh.Form {
	opts := make([]huh.Option[string], 0, len(worktrees))

	for _, wt := range worktrees {
		if wt.Bare {
			continue
		}
		name := filepath.Base(wt.Path)
		branch := wt.Branch
		if branch == "" && len(wt.Head) >= 7 {
			branch = fmt.Sprintf("detached:%s", wt.Head[:7])
		}

		label := fmt.Sprintf("%s (%s)", name, branch)
		opts = append(opts, huh.NewOption(label, name))
	}

	return huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select worktree").
				Options(opts...).
				Value(selected),
		),
	).WithTheme(Theme())
}

func StashConfirmForm(confirmed *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("You have uncommitted changes. Stash them?").
				Description("Changes will be restored when you return").
				Affirmative("Yes, stash").
				Negative("No, cancel").
				Value(confirmed),
		),
	).WithTheme(Theme())
}

// CleanupConfirmForm asks if user wants to delete worktree and branch after merge
func CleanupConfirmForm(worktreeName string, confirmed *bool) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete worktree '%s' and its branch?", worktreeName)).
				Affirmative("Yes, delete").
				Negative("No, keep").
				Value(confirmed),
		),
	).WithTheme(Theme())
}
