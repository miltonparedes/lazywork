package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Output handles dual-mode output (JSON for agents, styled for humans)
type Output struct {
	json    bool
	noColor bool
	isTTY   bool
	out     io.Writer
	errOut  io.Writer
	styles  *Styles
}

type Styles struct {
	Error   lipgloss.Style
	Success lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
	Dim     lipgloss.Style
	Bold    lipgloss.Style
}

func New(jsonFlag, noColorFlag bool) *Output {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))

	o := &Output{
		json:    jsonFlag,
		noColor: noColorFlag || !isTTY,
		isTTY:   isTTY,
		out:     os.Stdout,
		errOut:  os.Stderr,
	}

	// Note: lipgloss auto-detects color profile based on terminal
	// The noColor flag is handled by not using styles when printing

	o.styles = &Styles{
		Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true),
		Success: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
		Info:    lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		Dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Bold:    lipgloss.NewStyle().Bold(true),
	}

	return o
}

// IsTTY returns true if running interactively (not piped, not --json)
func (o *Output) IsTTY() bool {
	return o.isTTY && !o.json
}

func (o *Output) IsJSON() bool {
	return o.json
}

func (o *Output) JSON(v interface{}) error {
	enc := json.NewEncoder(o.out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (o *Output) Print(format string, args ...interface{}) {
	fmt.Fprintf(o.out, format, args...)
}

func (o *Output) Println(args ...interface{}) {
	fmt.Fprintln(o.out, args...)
}

func (o *Output) Success(msg string) {
	if o.json {
		return
	}
	text := "✓ " + msg
	if o.noColor {
		fmt.Fprintln(o.out, text)
	} else {
		fmt.Fprintln(o.out, o.styles.Success.Render(text))
	}
}

func (o *Output) Error(msg string) {
	text := "✗ " + msg
	if o.noColor {
		fmt.Fprintln(o.errOut, text)
	} else {
		fmt.Fprintln(o.errOut, o.styles.Error.Render(text))
	}
}

func (o *Output) Warning(msg string) {
	if o.json {
		return
	}
	text := "⚠ " + msg
	if o.noColor {
		fmt.Fprintln(o.errOut, text)
	} else {
		fmt.Fprintln(o.errOut, o.styles.Warning.Render(text))
	}
}

func (o *Output) Info(msg string) {
	if o.json {
		return
	}
	text := "ℹ " + msg
	if o.noColor {
		fmt.Fprintln(o.out, text)
	} else {
		fmt.Fprintln(o.out, o.styles.Info.Render(text))
	}
}

func (o *Output) Dim(msg string) {
	if o.json {
		return
	}
	if o.noColor {
		fmt.Fprintln(o.out, msg)
	} else {
		fmt.Fprintln(o.out, o.styles.Dim.Render(msg))
	}
}

func (o *Output) Bold(msg string) {
	if o.json {
		return
	}
	if o.noColor {
		fmt.Fprintln(o.out, msg)
	} else {
		fmt.Fprintln(o.out, o.styles.Bold.Render(msg))
	}
}

// Result handles dual-mode output - JSON data or human message
func (o *Output) Result(data interface{}, humanMsg string) {
	if o.json {
		o.JSON(data)
	} else {
		o.Print("%s\n", humanMsg)
	}
}

func (o *Output) ErrorResult(err error, code string) {
	if o.json {
		o.JSON(map[string]string{
			"error": err.Error(),
			"code":  code,
		})
	} else {
		o.Error(err.Error())
	}
}

func (o *Output) Styles() *Styles {
	return o.styles
}
