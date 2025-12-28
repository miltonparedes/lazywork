package tui

import (
	"github.com/charmbracelet/huh"
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
