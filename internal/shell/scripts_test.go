package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		shellEnv string
		expected string
	}{
		{"/bin/bash", Bash},
		{"/usr/bin/bash", Bash},
		{"/bin/zsh", Zsh},
		{"/usr/bin/zsh", Zsh},
		{"/usr/bin/fish", Fish},
		{"/opt/homebrew/bin/fish", Fish},
		{"", Bash}, // default
		{"/bin/sh", Bash},
	}

	for _, tt := range tests {
		t.Run(tt.shellEnv, func(t *testing.T) {
			orig := os.Getenv("SHELL")
			os.Setenv("SHELL", tt.shellEnv)
			defer os.Setenv("SHELL", orig)

			got := DetectShell()
			if got != tt.expected {
				t.Errorf("DetectShell() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsValidShell(t *testing.T) {
	valid := []string{"bash", "zsh", "fish"}
	for _, s := range valid {
		if !IsValidShell(s) {
			t.Errorf("IsValidShell(%q) = false, want true", s)
		}
	}

	invalid := []string{"powershell", "cmd", "sh", "tcsh", ""}
	for _, s := range invalid {
		if IsValidShell(s) {
			t.Errorf("IsValidShell(%q) = true, want false", s)
		}
	}
}

func TestSupportedShells(t *testing.T) {
	shells := SupportedShells()

	if len(shells) != 3 {
		t.Errorf("SupportedShells() returned %d shells, want 3", len(shells))
	}

	expected := map[string]bool{"bash": true, "zsh": true, "fish": true}
	for _, s := range shells {
		if !expected[s] {
			t.Errorf("unexpected shell %q in SupportedShells()", s)
		}
	}
}

func TestInitScript(t *testing.T) {
	tests := []struct {
		shell    string
		contains []string
	}{
		{
			Bash,
			[]string{"__lazywork_exec", "alias lw=", "alias lwt=", "--shell-helper"},
		},
		{
			Zsh,
			[]string{"__lazywork_exec", "alias lw=", "alias lwt=", "--shell-helper"},
		},
		{
			Fish,
			[]string{"function __lazywork_exec", "alias lw=", "alias lwt=", "--shell-helper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			script := InitScript(tt.shell)

			for _, substr := range tt.contains {
				if !strings.Contains(script, substr) {
					t.Errorf("InitScript(%q) missing %q", tt.shell, substr)
				}
			}
		})
	}
}

func TestInitScriptBashZshSimilar(t *testing.T) {
	bashScript := InitScript(Bash)
	zshScript := InitScript(Zsh)

	if !strings.Contains(bashScript, "__lazywork_exec()") {
		t.Error("Bash script missing __lazywork_exec() function")
	}
	if !strings.Contains(zshScript, "__lazywork_exec()") {
		t.Error("Zsh script missing __lazywork_exec() function")
	}
}

func TestInitScriptFishDifferent(t *testing.T) {
	fishScript := InitScript(Fish)

	// Fish uses different syntax
	if !strings.Contains(fishScript, "function __lazywork_exec") {
		t.Error("Fish script missing 'function __lazywork_exec'")
	}
	if !strings.Contains(fishScript, "set -l") {
		t.Error("Fish script missing 'set -l' (fish variable syntax)")
	}
	if !strings.Contains(fishScript, "test -s") {
		t.Error("Fish script missing 'test -s' (file size check)")
	}
}

func TestRcFile(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		shell    string
		expected string
	}{
		{Bash, filepath.Join(home, ".bashrc")},
		{Zsh, filepath.Join(home, ".zshrc")},
		{Fish, filepath.Join(home, ".config", "fish", "config.fish")},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := RcFile(tt.shell)
			if got != tt.expected {
				t.Errorf("RcFile(%q) = %q, want %q", tt.shell, got, tt.expected)
			}
		})
	}
}

func TestInitLine(t *testing.T) {
	tests := []struct {
		shell    string
		contains string
	}{
		{Bash, `eval "$(lazywork shell init bash)"`},
		{Zsh, `eval "$(lazywork shell init zsh)"`},
		{Fish, "lazywork shell init fish | source"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := InitLine(tt.shell)
			if got != tt.contains {
				t.Errorf("InitLine(%q) = %q, want %q", tt.shell, got, tt.contains)
			}
		})
	}
}

func TestInitScriptHandlesCd(t *testing.T) {
	for _, shell := range SupportedShells() {
		script := InitScript(shell)

		// All scripts use a cd_file for navigation
		if !strings.Contains(script, "cd_file") {
			t.Errorf("InitScript(%q) doesn't use cd_file", shell)
		}

		// All scripts source the cd file
		if !strings.Contains(script, "source") {
			t.Errorf("InitScript(%q) doesn't source cd file", shell)
		}
	}
}

func TestInitScriptPreservesExitCode(t *testing.T) {
	for _, shell := range SupportedShells() {
		script := InitScript(shell)

		if shell == Fish {
			if !strings.Contains(script, "$status") {
				t.Errorf("Fish script doesn't capture $status")
			}
		} else {
			if !strings.Contains(script, "$?") || !strings.Contains(script, "exit_code") {
				t.Errorf("%s script doesn't capture exit code", shell)
			}
		}
	}
}
