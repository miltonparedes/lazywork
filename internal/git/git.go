package git

import (
	"bytes"
	"fmt"
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
			if strings.HasPrefix(branch, "refs/heads/") {
				branch = strings.TrimPrefix(branch, "refs/heads/")
			}
			current.Branch = branch
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
