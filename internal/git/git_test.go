package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testRepo creates a temporary git repository for testing
type testRepo struct {
	t    *testing.T
	dir  string
	orig string
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()

	dir, err := os.MkdirTemp("", "lazywork-test-*")
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

	// Initialize git repo
	if err := runCmd("git", "init"); err != nil {
		os.Chdir(orig)
		os.RemoveAll(dir)
		t.Fatalf("failed to init git: %v", err)
	}

	// Configure git user for commits
	runCmd("git", "config", "user.email", "test@test.com")
	runCmd("git", "config", "user.name", "Test User")

	// Create initial commit
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

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// Test basic git detection
func TestIsInsideWorkTree(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	if !IsInsideWorkTree() {
		t.Error("expected to be inside work tree")
	}
}

func TestGetRepoRoot(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	root, err := GetRepoRoot()
	if err != nil {
		t.Fatalf("GetRepoRoot failed: %v", err)
	}

	// Should match the temp directory
	if root != repo.dir {
		t.Errorf("expected root=%s, got=%s", repo.dir, root)
	}
}

func TestCurrentBranch(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	branch, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch failed: %v", err)
	}

	// Default branch could be main or master
	if branch != "main" && branch != "master" {
		t.Errorf("expected main or master, got=%s", branch)
	}
}

// Test uncommitted changes detection
func TestHasUncommittedChanges(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Clean state - should have no changes
	if HasUncommittedChanges() {
		t.Error("expected no uncommitted changes in clean repo")
	}

	// Create a new file (untracked)
	os.WriteFile("newfile.txt", []byte("test"), 0o644)
	if !HasUncommittedChanges() {
		t.Error("expected uncommitted changes after adding file")
	}

	// Stage it
	runCmd("git", "add", "newfile.txt")
	if !HasUncommittedChanges() {
		t.Error("expected uncommitted changes with staged file")
	}

	// Commit it
	runCmd("git", "commit", "-m", "add file")
	if HasUncommittedChanges() {
		t.Error("expected no uncommitted changes after commit")
	}
}

// Test checkout
func TestCheckout(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create a new branch
	runCmd("git", "branch", "feature-test")

	// Checkout the new branch
	if err := Checkout("feature-test"); err != nil {
		t.Fatalf("Checkout failed: %v", err)
	}

	branch, _ := CurrentBranch()
	if branch != "feature-test" {
		t.Errorf("expected branch=feature-test, got=%s", branch)
	}
}

// Test stash operations
func TestStash(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create uncommitted changes
	os.WriteFile("test.txt", []byte("changes"), 0o644)
	runCmd("git", "add", "test.txt")

	if !HasUncommittedChanges() {
		t.Fatal("expected uncommitted changes")
	}

	// Stash the changes
	ref, err := Stash("test stash")
	if err != nil {
		t.Fatalf("Stash failed: %v", err)
	}

	if !strings.HasPrefix(ref, "stash@{") {
		t.Errorf("expected stash ref, got=%s", ref)
	}

	// Should be clean now
	if HasUncommittedChanges() {
		t.Error("expected no uncommitted changes after stash")
	}

	// Pop the stash
	if err := StashPop(); err != nil {
		t.Fatalf("StashPop failed: %v", err)
	}

	// Changes should be back
	if !HasUncommittedChanges() {
		t.Error("expected uncommitted changes after stash pop")
	}
}

// Test state management
func TestStateManagement(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	key := "LAZYWORK_TEST_STATE"
	value := "test-value-123"

	// Save state
	if err := SaveState(key, value); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// Load state
	loaded, err := LoadState(key)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if loaded != value {
		t.Errorf("expected value=%s, got=%s", value, loaded)
	}

	// Clear state
	if err := ClearState(key); err != nil {
		t.Fatalf("ClearState failed: %v", err)
	}

	// Should fail to load now
	_, err = LoadState(key)
	if err == nil {
		t.Error("expected error loading cleared state")
	}
}

// Test use state helpers
func TestUseState(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Initially no saved state
	if HasSavedState() {
		t.Error("expected no saved state initially")
	}

	// Save use state
	if err := SaveUseState("main", "stash@{0}"); err != nil {
		t.Fatalf("SaveUseState failed: %v", err)
	}

	// Should have state now
	if !HasSavedState() {
		t.Error("expected saved state after SaveUseState")
	}

	// Load and verify
	branch, stashRef, err := LoadUseState()
	if err != nil {
		t.Fatalf("LoadUseState failed: %v", err)
	}
	if branch != "main" {
		t.Errorf("expected branch=main, got=%s", branch)
	}
	if stashRef != "stash@{0}" {
		t.Errorf("expected stashRef=stash@{0}, got=%s", stashRef)
	}

	// Clear
	if err := ClearUseState(); err != nil {
		t.Fatalf("ClearUseState failed: %v", err)
	}

	if HasSavedState() {
		t.Error("expected no saved state after clear")
	}
}

// Test use state without stash
func TestUseStateNoStash(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Save without stash
	if err := SaveUseState("feature", ""); err != nil {
		t.Fatalf("SaveUseState failed: %v", err)
	}

	branch, stashRef, err := LoadUseState()
	if err != nil {
		t.Fatalf("LoadUseState failed: %v", err)
	}
	if branch != "feature" {
		t.Errorf("expected branch=feature, got=%s", branch)
	}
	if stashRef != "" {
		t.Errorf("expected empty stashRef, got=%s", stashRef)
	}
}

// Test branch operations
func TestBranchExists(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	mainBranch := GetMainBranch()
	if !BranchExists(mainBranch) {
		t.Errorf("expected %s branch to exist", mainBranch)
	}

	if BranchExists("nonexistent-branch-xyz") {
		t.Error("expected nonexistent branch to not exist")
	}
}

func TestDeleteBranch(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create and checkout a branch
	runCmd("git", "checkout", "-b", "to-delete")
	runCmd("git", "checkout", "-") // go back to main

	if !BranchExists("to-delete") {
		t.Fatal("branch should exist before delete")
	}

	if err := DeleteBranch("to-delete", false); err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	if BranchExists("to-delete") {
		t.Error("branch should not exist after delete")
	}
}

// Test merge
func TestMerge(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	mainBranch := GetMainBranch()

	// Create a feature branch with changes
	runCmd("git", "checkout", "-b", "feature-merge")
	os.WriteFile("feature.txt", []byte("feature content"), 0o644)
	runCmd("git", "add", "feature.txt")
	runCmd("git", "commit", "-m", "add feature")

	// Go back to main
	runCmd("git", "checkout", mainBranch)

	// Merge feature
	if err := Merge("feature-merge"); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// File should exist in main now
	if _, err := os.Stat("feature.txt"); os.IsNotExist(err) {
		t.Error("expected feature.txt to exist after merge")
	}
}

// Test worktree operations
func TestListWorktrees(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	worktrees, err := ListWorktrees()
	if err != nil {
		t.Fatalf("ListWorktrees failed: %v", err)
	}

	// Should have at least the main worktree
	if len(worktrees) < 1 {
		t.Error("expected at least 1 worktree")
	}
}

func TestFindWorktreeByName(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// Create a worktree
	wtPath := filepath.Join(repo.dir, ".worktrees", "test-feature")
	if err := AddWorktree(wtPath, "test-feature"); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Find by name
	wt, err := FindWorktreeByName("test-feature")
	if err != nil {
		t.Fatalf("FindWorktreeByName failed: %v", err)
	}

	if wt.Branch != "test-feature" {
		t.Errorf("expected branch=test-feature, got=%s", wt.Branch)
	}

	// Should not find nonexistent
	_, err = FindWorktreeByName("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent worktree")
	}
}

// Test IsMainWorktree
func TestIsMainWorktree(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	// In main repo, should be true
	if !IsMainWorktree() {
		t.Error("expected to be in main worktree")
	}

	// Create and go to a secondary worktree
	wtPath := filepath.Join(repo.dir, ".worktrees", "secondary")
	if err := AddWorktree(wtPath, "secondary"); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	if err := os.Chdir(wtPath); err != nil {
		t.Fatalf("failed to chdir to worktree: %v", err)
	}

	// In secondary worktree, should be false
	if IsMainWorktree() {
		t.Error("expected NOT to be in main worktree")
	}
}

func TestGetGitDir(t *testing.T) {
	repo := newTestRepo(t)
	defer repo.cleanup()

	gitDir, err := GetGitDir()
	if err != nil {
		t.Fatalf("GetGitDir failed: %v", err)
	}

	expected := filepath.Join(repo.dir, ".git")
	if gitDir != expected {
		t.Errorf("expected gitDir=%s, got=%s", expected, gitDir)
	}
}
