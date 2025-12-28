package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Path   string `json:"path"`
	Head   string `json:"head"`
	Branch string `json:"branch,omitempty"`
	Bare   bool   `json:"bare,omitempty"`
}

func IsInsideWorkTree() bool {
	_, err := runGit("rev-parse", "--is-inside-work-tree")
	return err == nil
}

func GetRepoRoot() (string, error) {
	output, err := runGit("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func CurrentBranch() (string, error) {
	output, err := runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func ListWorktrees() ([]Worktree, error) {
	output, err := runGit("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var current *Worktree

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				worktrees = append(worktrees, *current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current = &Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.Head = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			branch := strings.TrimPrefix(line, "branch ")
			// Extract branch name from refs/heads/...
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		} else if line == "bare" && current != nil {
			current.Bare = true
		}
	}

	// Don't forget the last worktree
	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees, nil
}

func AddWorktree(path, branch string) error {
	_, err := runGit("worktree", "add", path, "-b", branch)
	return err
}

func AddWorktreeFromBranch(path, branch string) error {
	_, err := runGit("worktree", "add", path, branch)
	return err
}

func RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	_, err := runGit(args...)
	return err
}

func PruneWorktrees() error {
	_, err := runGit("worktree", "prune")
	return err
}

func GetStagedDiff() (string, error) {
	return runGit("diff", "--staged")
}

func GetUnstagedDiff() (string, error) {
	return runGit("diff")
}

func Commit(message string) error {
	_, err := runGit("commit", "-m", message)
	return err
}

func BranchExists(name string) bool {
	_, err := runGit("rev-parse", "--verify", "refs/heads/"+name)
	return err == nil
}

// baseDir is relative to repo root (e.g., ".worktrees")
func GetWorktreePath(baseDir, name string) (string, error) {
	root, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, baseDir, name), nil
}

func HasUncommittedChanges() bool {
	output, err := runGit("status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) != ""
}

func Checkout(branch string) error {
	_, err := runGit("checkout", branch)
	return err
}

// Stash saves uncommitted changes and returns the stash reference
func Stash(message string) (string, error) {
	args := []string{"stash", "push"}
	if message != "" {
		args = append(args, "-m", message)
	}
	_, err := runGit(args...)
	if err != nil {
		return "", err
	}
	// Get the stash reference
	output, err := runGit("stash", "list", "-1")
	if err != nil {
		return "", err
	}
	// Extract stash@{0} or similar
	parts := strings.SplitN(strings.TrimSpace(output), ":", 2)
	if len(parts) > 0 {
		return parts[0], nil
	}
	return "stash@{0}", nil
}

func StashPop() error {
	_, err := runGit("stash", "pop")
	return err
}

func Merge(branch string) error {
	_, err := runGit("merge", branch)
	return err
}

func DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := runGit("branch", flag, name)
	return err
}

// GetMainBranch returns "main" or "master" depending on what exists
func GetMainBranch() string {
	if BranchExists("main") {
		return "main"
	}
	return "master"
}

func GetGitDir() (string, error) {
	output, err := runGit("rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	path := strings.TrimSpace(output)
	// Make absolute if relative
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(cwd, path)
	}
	return path, nil
}

// IsMainWorktree returns true if we're in the main worktree (not a secondary worktree)
func IsMainWorktree() bool {
	gitDir, err := GetGitDir()
	if err != nil {
		return false
	}
	// In secondary worktrees, git-dir contains /worktrees/ path
	// e.g., /path/to/repo/.git/worktrees/feature-name
	return !strings.Contains(gitDir, string(filepath.Separator)+"worktrees"+string(filepath.Separator))
}

const (
	statePreviousBranch = "LAZYWORK_PREVIOUS_BRANCH"
	stateStashRef       = "LAZYWORK_STASH_REF"
)

func SaveState(key, value string) error {
	gitDir, err := GetGitDir()
	if err != nil {
		return err
	}
	path := filepath.Join(gitDir, key)
	return os.WriteFile(path, []byte(value), 0o644)
}

func LoadState(key string) (string, error) {
	gitDir, err := GetGitDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(gitDir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func ClearState(key string) error {
	gitDir, err := GetGitDir()
	if err != nil {
		return err
	}
	path := filepath.Join(gitDir, key)
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// HasSavedState returns true if there's saved state from a previous 'use' command
func HasSavedState() bool {
	_, err := LoadState(statePreviousBranch)
	return err == nil
}

// SaveUseState saves the current branch and optional stash ref for later return
func SaveUseState(previousBranch, stashRef string) error {
	if err := SaveState(statePreviousBranch, previousBranch); err != nil {
		return err
	}
	if stashRef != "" {
		return SaveState(stateStashRef, stashRef)
	}
	return nil
}

func LoadUseState() (previousBranch, stashRef string, err error) {
	previousBranch, err = LoadState(statePreviousBranch)
	if err != nil {
		return "", "", err
	}
	stashRef, _ = LoadState(stateStashRef)
	return previousBranch, stashRef, nil
}

// ClearUseState removes all saved state from a 'use' command
func ClearUseState() error {
	if err := ClearState(statePreviousBranch); err != nil {
		return err
	}
	return ClearState(stateStashRef)
}

// FindWorktreeByName finds a worktree by name (basename match)
func FindWorktreeByName(name string) (*Worktree, error) {
	worktrees, err := ListWorktrees()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Bare {
			continue
		}
		// Match by basename or branch name
		if filepath.Base(wt.Path) == name || wt.Branch == name {
			return &wt, nil
		}
	}
	return nil, fmt.Errorf("worktree '%s' not found", name)
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), errMsg)
	}

	return stdout.String(), nil
}
