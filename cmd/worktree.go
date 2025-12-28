package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/miltonparedes/lazywork/internal/git"
	"github.com/miltonparedes/lazywork/internal/output"
	"github.com/miltonparedes/lazywork/internal/tui"
	"github.com/miltonparedes/lazywork/pkg/config"
	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage git worktrees",
	Long:  "List, create, and manage git worktrees with AI-powered naming.",
}

var worktreeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all worktrees",
	RunE:  runWorktreeList,
}

var worktreeAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Create a new worktree",
	Long: `Create a new worktree with the specified name.
The worktree will be created in .worktrees/<name> by default.

If no name is provided, you'll be prompted to enter one interactively.

Example:
  lazywork worktree add feature-auth
  # Creates .worktrees/feature-auth with branch feature-auth

  lazywork worktree add
  # Prompts for branch name interactively`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWorktreeAdd,
}

var worktreeRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Args:    cobra.ExactArgs(1),
	RunE:    runWorktreeRemove,
}

var worktreePruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stale worktree entries",
	RunE:  runWorktreePrune,
}

var (
	forceRemove bool
	fromBranch  string
)

func init() {
	rootCmd.AddCommand(worktreeCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	worktreeCmd.AddCommand(worktreeAddCmd)
	worktreeCmd.AddCommand(worktreeRemoveCmd)
	worktreeCmd.AddCommand(worktreePruneCmd)

	worktreeRemoveCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force removal even with uncommitted changes")
	worktreeAddCmd.Flags().StringVarP(&fromBranch, "branch", "b", "", "Create worktree from existing branch instead of new branch")
}

func runWorktreeList(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		out.ErrorResult(err, "WORKTREE_LIST_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"worktrees": worktrees,
			"count":     len(worktrees),
		})
	}

	if len(worktrees) == 0 {
		out.Dim("No worktrees found")
		return nil
	}

	out.Bold(fmt.Sprintf("Worktrees (%d):", len(worktrees)))
	out.Println()

	for _, wt := range worktrees {
		if wt.Bare {
			out.Print("  %s (bare)\n", wt.Path)
		} else {
			branch := wt.Branch
			if branch == "" {
				branch = fmt.Sprintf("(detached at %s)", wt.Head[:7])
			}
			out.Print("  %s\n", filepath.Base(wt.Path))
			out.Dim(fmt.Sprintf("    branch: %s", branch))
			out.Dim(fmt.Sprintf("    path:   %s", wt.Path))
		}
		out.Println()
	}

	return nil
}

func runWorktreeAdd(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	cfg, err := config.LoadFrom(cfgFile)
	if err != nil {
		out.ErrorResult(err, "CONFIG_LOAD_ERROR")
		return err
	}

	var name string
	if len(args) > 0 {
		name = args[0]
	} else if out.IsTTY() {
		form := tui.BranchNameForm(&name)
		if err := form.Run(); err != nil {
			return err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			err := fmt.Errorf("branch name cannot be empty")
			out.ErrorResult(err, "EMPTY_NAME")
			return err
		}
	} else {
		err := fmt.Errorf("branch name required (use: lazywork worktree add <name>)")
		out.ErrorResult(err, "NAME_REQUIRED")
		return err
	}

	worktreePath, err := git.GetWorktreePath(cfg.GetWorktreeDir(), name)
	if err != nil {
		out.ErrorResult(err, "PATH_ERROR")
		return err
	}

	var branch string
	if fromBranch != "" {
		// Use existing branch
		if !git.BranchExists(fromBranch) {
			err := fmt.Errorf("branch '%s' does not exist", fromBranch)
			out.ErrorResult(err, "BRANCH_NOT_FOUND")
			return err
		}
		branch = fromBranch
		err = git.AddWorktreeFromBranch(worktreePath, branch)
	} else {
		// Create new branch
		branch = name
		if git.BranchExists(branch) {
			err := fmt.Errorf("branch '%s' already exists. Use --branch to checkout existing branch", branch)
			out.ErrorResult(err, "BRANCH_EXISTS")
			return err
		}
		err = git.AddWorktree(worktreePath, branch)
	}

	if err != nil {
		out.ErrorResult(err, "WORKTREE_ADD_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"path":    worktreePath,
			"branch":  branch,
			"created": true,
		})
	}

	out.Success(fmt.Sprintf("Created worktree: %s", name))
	out.Dim(fmt.Sprintf("  branch: %s", branch))
	out.Dim(fmt.Sprintf("  path:   %s", worktreePath))
	out.Println()
	out.Info(fmt.Sprintf("cd %s", worktreePath))

	return nil
}

func runWorktreeRemove(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)
	name := args[0]

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		out.ErrorResult(err, "WORKTREE_LIST_ERROR")
		return err
	}

	var targetPath string
	for _, wt := range worktrees {
		// Match by name (basename of path) or full path
		if filepath.Base(wt.Path) == name || wt.Path == name {
			targetPath = wt.Path
			break
		}
		// Also match by suffix pattern (repo-name)
		if matched, _ := filepath.Match("*-"+name, filepath.Base(wt.Path)); matched {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		err := fmt.Errorf("worktree '%s' not found", name)
		out.ErrorResult(err, "WORKTREE_NOT_FOUND")
		return err
	}

	if err := git.RemoveWorktree(targetPath, forceRemove); err != nil {
		out.ErrorResult(err, "WORKTREE_REMOVE_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"path":    targetPath,
			"removed": true,
		})
	}

	out.Success(fmt.Sprintf("Removed worktree: %s", filepath.Base(targetPath)))

	return nil
}

func runWorktreePrune(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	if err := git.PruneWorktrees(); err != nil {
		out.ErrorResult(err, "WORKTREE_PRUNE_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"pruned": true,
		})
	}

	out.Success("Pruned stale worktree entries")

	return nil
}
