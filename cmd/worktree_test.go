package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/miltonparedes/lazywork/internal/git"
	"github.com/miltonparedes/lazywork/internal/output"
	"github.com/miltonparedes/lazywork/internal/state"
	"github.com/spf13/cobra"
)

type testRepo struct {
	t    *testing.T
	dir  string
	orig string
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()

	dir, err := os.MkdirTemp("", "lazywork-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	orig, err := os.Getwd()
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to get current dir: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to chdir to temp: %v", err)
	}

	if err := runCmd("git", "init"); err != nil {
		os.Chdir(orig)
		os.RemoveAll(dir)
		t.Fatalf("failed to init git: %v", err)
	}

	runCmd("git", "config", "user.email", "test@test.com")
	runCmd("git", "config", "user.name", "Test User")

	if err := os.WriteFile("README.md", []byte("# Test\n"), 0o644); err != nil {
		os.Chdir(orig)
		os.RemoveAll(dir)
		t.Fatalf("failed to create file: %v", err)
	}
	runCmd("git", "add", ".")
	runCmd("git", "commit", "-m", "Initial commit")

	return &testRepo{t: t, dir: dir, orig: orig}
}

func (r *testRepo) cleanup() {
	os.Chdir(r.orig)
	os.RemoveAll(r.dir)
}

func (r *testRepo) createWorktree(name string) string {
	r.t.Helper()
	worktreesDir := filepath.Join(r.dir, ".worktrees")
	if err := os.MkdirAll(worktreesDir, 0o755); err != nil {
		r.t.Fatalf("failed to create worktrees dir: %v", err)
	}

	wtPath := filepath.Join(worktreesDir, name)
	if err := runCmd("git", "worktree", "add", "-b", name, wtPath); err != nil {
		r.t.Fatalf("failed to create worktree: %v", err)
	}
	return wtPath
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func TestCompleteWorktreeNamesEmpty(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	completions, directive := completeWorktreeNames(&cobra.Command{}, []string{}, "")

	if len(completions) != 0 {
		t.Errorf("expected no completions, got %v", completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
}

func TestCompleteWorktreeNamesWithWorktrees(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	repo.createWorktree("feature-auth")
	repo.createWorktree("feature-api")

	completions, directive := completeWorktreeNames(&cobra.Command{}, []string{}, "")

	if len(completions) != 2 {
		t.Errorf("expected 2 completions, got %d: %v", len(completions), completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
}

func TestCompleteWorktreeNamesFiltering(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	repo.createWorktree("feature-auth")
	repo.createWorktree("bugfix-login")

	completions, _ := completeWorktreeNames(&cobra.Command{}, []string{}, "feat")

	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d: %v", len(completions), completions)
	}
}

func TestCompleteWorktreeNamesAlreadyHasArg(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	repo.createWorktree("feature-auth")

	completions, directive := completeWorktreeNames(&cobra.Command{}, []string{"feature-auth"}, "")

	if len(completions) != 0 {
		t.Errorf("expected no completions when arg already provided, got %v", completions)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
}

func TestHistoryRecordVisitOnNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	repo := newTestRepo(t)
	defer repo.cleanup()

	wt1 := repo.createWorktree("feature-a")
	wt2 := repo.createWorktree("feature-b")

	history, err := state.LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	history.RecordVisit(repo.dir, wt1)
	history.RecordVisit(repo.dir, wt2)

	if err := history.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := state.LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	previous := loaded.GetPrevious(repo.dir)
	if previous != wt1 {
		t.Errorf("GetPrevious() = %q, want %q", previous, wt1)
	}

	current := loaded.GetCurrent(repo.dir)
	if current != wt2 {
		t.Errorf("GetCurrent() = %q, want %q", current, wt2)
	}
}

func TestPreviousFlagDefined(t *testing.T) {
	flag := worktreeGoCmd.Flags().Lookup("previous")
	if flag == nil {
		t.Error("--previous flag not defined on worktreeGoCmd")
	}

	shortFlag := worktreeGoCmd.Flags().ShorthandLookup("p")
	if shortFlag == nil {
		t.Error("-p shorthand not defined on worktreeGoCmd")
	}
}

func TestWorktreeCmdHasRunE(t *testing.T) {
	if worktreeCmd.RunE == nil {
		t.Error("worktreeCmd.RunE should be defined for interactive selector")
	}
}

func TestValidArgsFunctionsDefined(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"worktreeGoCmd", worktreeGoCmd},
		{"worktreeRemoveCmd", worktreeRemoveCmd},
		{"worktreeUseCmd", worktreeUseCmd},
		{"worktreeFinishCmd", worktreeFinishCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.ValidArgsFunction == nil {
				t.Errorf("%s should have ValidArgsFunction defined", tt.name)
			}
		})
	}
}

func TestNavigateToWorktreeRecordsHistory(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	repo := newTestRepo(t)
	defer repo.cleanup()

	wt1 := repo.createWorktree("feature-a")
	wt2 := repo.createWorktree("feature-b")

	out := newTestOutput()
	if err := navigateToWorktree(out, wt1); err != nil {
		t.Fatalf("navigateToWorktree failed: %v", err)
	}

	if err := navigateToWorktree(out, wt2); err != nil {
		t.Fatalf("navigateToWorktree failed: %v", err)
	}

	history, err := state.LoadHistory()
	if err != nil {
		t.Fatalf("LoadHistory failed: %v", err)
	}

	// History is keyed by git common dir, not repo root
	repoKey, err := git.GetGitCommonDir()
	if err != nil {
		t.Fatalf("GetGitCommonDir failed: %v", err)
	}

	if history.GetCurrent(repoKey) != wt2 {
		t.Errorf("current = %q, want %q", history.GetCurrent(repoKey), wt2)
	}

	if history.GetPrevious(repoKey) != wt1 {
		t.Errorf("previous = %q, want %q", history.GetPrevious(repoKey), wt1)
	}
}

func TestNavigateToWorktreeShellHelper(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	wt := repo.createWorktree("feature-test")

	originalShellHelper := shellHelper
	shellHelper = true
	defer func() { shellHelper = originalShellHelper }()

	out := newTestOutput()
	if err := navigateToWorktree(out, wt); err != nil {
		t.Fatalf("navigateToWorktree failed: %v", err)
	}

	// With shellHelper, function should return nil (prints cd command to stdout)
	// We can't easily capture stdout here, but we verify no error
}

func TestGoToPreviousWorktreeNoPrevious(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	repo := newTestRepo(t)
	defer repo.cleanup()

	repo.createWorktree("feature-a")

	out := newTestOutput()

	err := goToPreviousWorktree(out)
	if err == nil {
		t.Error("expected error when no previous worktree and not TTY")
	}
}

func TestGoToPreviousWorktreeWithPrevious(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	repo := newTestRepo(t)
	defer repo.cleanup()

	wt1 := repo.createWorktree("feature-a")
	wt2 := repo.createWorktree("feature-b")

	// History is keyed by git common dir
	repoKey, err := git.GetGitCommonDir()
	if err != nil {
		t.Fatalf("GetGitCommonDir failed: %v", err)
	}

	history, _ := state.LoadHistory()
	history.RecordVisit(repoKey, wt1)
	history.RecordVisit(repoKey, wt2)
	history.Save()

	out := newTestOutput()

	err = goToPreviousWorktree(out)
	if err != nil {
		t.Fatalf("goToPreviousWorktree failed: %v", err)
	}

	history, _ = state.LoadHistory()
	if history.GetCurrent(repoKey) != wt1 {
		t.Errorf("current = %q, want %q", history.GetCurrent(repoKey), wt1)
	}
}

// TestCrossWorktreeNavigation verifies that history works across worktrees
// This is the critical fix: history should be keyed by git common dir, not repo root
func TestCrossWorktreeNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	repo := newTestRepo(t)
	defer repo.cleanup()

	wt1 := repo.createWorktree("feature-a")
	wt2 := repo.createWorktree("feature-b")

	out := newTestOutput()
	if err := navigateToWorktree(out, wt1); err != nil {
		t.Fatalf("navigateToWorktree failed: %v", err)
	}

	if err := navigateToWorktree(out, wt2); err != nil {
		t.Fatalf("navigateToWorktree failed: %v", err)
	}

	if err := os.Chdir(wt2); err != nil {
		t.Fatalf("failed to chdir to worktree: %v", err)
	}

	commonDirFromWT, err := git.GetGitCommonDir()
	if err != nil {
		t.Fatalf("GetGitCommonDir failed from worktree: %v", err)
	}

	os.Chdir(repo.dir)
	commonDirFromMain, err := git.GetGitCommonDir()
	if err != nil {
		t.Fatalf("GetGitCommonDir failed from main: %v", err)
	}

	if commonDirFromWT != commonDirFromMain {
		t.Errorf("GetGitCommonDir differs between worktrees: wt=%q main=%q", commonDirFromWT, commonDirFromMain)
	}

	os.Chdir(wt2)
	err = goToPreviousWorktree(out)
	if err != nil {
		t.Fatalf("goToPreviousWorktree from worktree failed: %v", err)
	}

	history, _ := state.LoadHistory()
	if history.GetCurrent(commonDirFromMain) != wt1 {
		t.Errorf("after go -, current = %q, want %q", history.GetCurrent(commonDirFromMain), wt1)
	}
}

// TestFirstNavigationRecordsOrigin verifies that the first navigation records
// the starting location, so lwt go - can return to it
func TestFirstNavigationRecordsOrigin(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	repo := newTestRepo(t)
	defer repo.cleanup()

	mainDir := repo.dir
	wt1 := repo.createWorktree("feature-a")

	repoKey, err := git.GetGitCommonDir()
	if err != nil {
		t.Fatalf("GetGitCommonDir failed: %v", err)
	}

	// Verify history is empty initially
	history, _ := state.LoadHistory()
	if history.GetCurrent(repoKey) != "" {
		t.Errorf("expected empty current initially, got %q", history.GetCurrent(repoKey))
	}

	// First navigation: from main to wt1
	out := newTestOutput()
	if err := navigateToWorktree(out, wt1); err != nil {
		t.Fatalf("navigateToWorktree failed: %v", err)
	}

	// Verify: origin (main) should be recorded as previous
	history, _ = state.LoadHistory()
	if history.GetCurrent(repoKey) != wt1 {
		t.Errorf("current = %q, want %q", history.GetCurrent(repoKey), wt1)
	}
	if history.GetPrevious(repoKey) != mainDir {
		t.Errorf("previous = %q, want %q (main dir)", history.GetPrevious(repoKey), mainDir)
	}

	// Now lwt go - should work and return to main
	err = goToPreviousWorktree(out)
	if err != nil {
		t.Fatalf("goToPreviousWorktree failed: %v", err)
	}

	history, _ = state.LoadHistory()
	if history.GetCurrent(repoKey) != mainDir {
		t.Errorf("after go -, current = %q, want %q", history.GetCurrent(repoKey), mainDir)
	}
}

func newTestOutput() *output.Output {
	return output.New(false, true)
}
