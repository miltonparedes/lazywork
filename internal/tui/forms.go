package tui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/miltonparedes/lazywork/internal/git"
)

func Theme() *huh.Theme {
	return huh.ThemeBase()
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
