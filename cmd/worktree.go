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

var worktreeGoCmd = &cobra.Command{
	Use:   "go [name]",
	Short: "Navigate to a worktree directory",
	Long: `Navigate to a worktree directory.

If no name is provided, you'll be prompted to select one interactively.

Setup shell integration for automatic cd:
  # Bash/Zsh
  eval "$(lazywork shell init)"

  # Fish
  lazywork shell init fish | source

Then use: lwt go [name]`,
	Aliases: []string{"cd"},
	Args:    cobra.MaximumNArgs(1),
	RunE:    runWorktreeGo,
}

var worktreeUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Checkout worktree branch in main repo",
	Long: `Checkout a worktree's branch in the main repository.

This is useful when your project configuration (Docker, databases, etc.)
only works in the main repository directory.

The command will:
1. Stash any uncommitted changes (with your permission)
2. Checkout the worktree's branch
3. Save state so you can return later with 'worktree return'`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWorktreeUse,
}

var worktreeReturnCmd = &cobra.Command{
	Use:   "return",
	Short: "Return to previous branch after 'use'",
	Long: `Return to the branch you were on before using 'worktree use'.

This will:
1. Checkout the previous branch
2. Restore any stashed changes`,
	RunE: runWorktreeReturn,
}

var worktreeFinishCmd = &cobra.Command{
	Use:   "finish [name]",
	Short: "Merge worktree branch and cleanup",
	Long: `Merge a worktree's branch into the current branch and optionally clean up.

This command must be run from the main branch (main/master).
After a successful merge, you'll be asked if you want to delete
the worktree and its branch.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWorktreeFinish,
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
	worktreeCmd.AddCommand(worktreeGoCmd)
	worktreeCmd.AddCommand(worktreeUseCmd)
	worktreeCmd.AddCommand(worktreeReturnCmd)
	worktreeCmd.AddCommand(worktreeFinishCmd)

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

func runWorktreeGo(cmd *cobra.Command, args []string) error {
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

	var secondaryWorktrees []git.Worktree
	for _, wt := range worktrees {
		if !wt.Bare && strings.Contains(wt.Path, string(filepath.Separator)+".worktrees"+string(filepath.Separator)) {
			secondaryWorktrees = append(secondaryWorktrees, wt)
		}
	}

	if len(secondaryWorktrees) == 0 {
		err := fmt.Errorf("no worktrees found. Create one with: lazywork worktree add <name>")
		out.ErrorResult(err, "NO_WORKTREES")
		return err
	}

	var name string
	if len(args) > 0 {
		name = args[0]
	} else if out.IsTTY() {
		form := tui.WorktreeSelectForm(secondaryWorktrees, &name)
		if err := form.Run(); err != nil {
			return err
		}
	} else {
		err := fmt.Errorf("worktree name required (use: lazywork worktree go <name>)")
		out.ErrorResult(err, "NAME_REQUIRED")
		return err
	}

	var targetPath string
	for _, wt := range secondaryWorktrees {
		if filepath.Base(wt.Path) == name || wt.Branch == name {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		err := fmt.Errorf("worktree '%s' not found", name)
		out.ErrorResult(err, "WORKTREE_NOT_FOUND")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"path": targetPath,
			"cd":   fmt.Sprintf("cd %s", targetPath),
		})
	}

	if shellHelper {
		fmt.Printf("cd %s\n", targetPath)
		return nil
	}

	out.Info(fmt.Sprintf("Run: cd %s", targetPath))
	out.Dim("Tip: Use 'lwt go' with shell integration for automatic cd")
	out.Dim("Setup: eval \"$(lazywork shell init)\"")

	return nil
}

func runWorktreeUse(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	if !git.IsMainWorktree() {
		err := fmt.Errorf("must be in main repository, not a worktree")
		out.ErrorResult(err, "NOT_MAIN_WORKTREE")
		return err
	}

	if git.HasSavedState() {
		err := fmt.Errorf("already using a worktree branch. Run 'lazywork worktree return' first")
		out.ErrorResult(err, "STATE_EXISTS")
		return err
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		out.ErrorResult(err, "WORKTREE_LIST_ERROR")
		return err
	}

	var secondaryWorktrees []git.Worktree
	for _, wt := range worktrees {
		if !wt.Bare && strings.Contains(wt.Path, string(filepath.Separator)+".worktrees"+string(filepath.Separator)) {
			secondaryWorktrees = append(secondaryWorktrees, wt)
		}
	}

	if len(secondaryWorktrees) == 0 {
		err := fmt.Errorf("no worktrees found")
		out.ErrorResult(err, "NO_WORKTREES")
		return err
	}

	var name string
	if len(args) > 0 {
		name = args[0]
	} else if out.IsTTY() {
		form := tui.WorktreeSelectForm(secondaryWorktrees, &name)
		if err := form.Run(); err != nil {
			return err
		}
	} else {
		err := fmt.Errorf("worktree name required")
		out.ErrorResult(err, "NAME_REQUIRED")
		return err
	}

	var targetWorktree *git.Worktree
	for _, wt := range secondaryWorktrees {
		if filepath.Base(wt.Path) == name || wt.Branch == name {
			targetWorktree = &wt
			break
		}
	}

	if targetWorktree == nil {
		err := fmt.Errorf("worktree '%s' not found", name)
		out.ErrorResult(err, "WORKTREE_NOT_FOUND")
		return err
	}

	if targetWorktree.Branch == "" {
		err := fmt.Errorf("worktree is in detached HEAD state")
		out.ErrorResult(err, "DETACHED_HEAD")
		return err
	}

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		out.ErrorResult(err, "BRANCH_ERROR")
		return err
	}

	var stashRef string
	if git.HasUncommittedChanges() {
		if out.IsTTY() {
			var doStash bool
			form := tui.StashConfirmForm(&doStash)
			if err := form.Run(); err != nil {
				return err
			}
			if !doStash {
				err := fmt.Errorf("cancelled: uncommitted changes would be lost")
				out.ErrorResult(err, "CANCELLED")
				return err
			}
		} else if !jsonOutput {
			err := fmt.Errorf("uncommitted changes detected. Commit or stash them first")
			out.ErrorResult(err, "UNCOMMITTED_CHANGES")
			return err
		}

		stashRef, err = git.Stash("lazywork: auto-stash before worktree use")
		if err != nil {
			out.ErrorResult(err, "STASH_ERROR")
			return err
		}
	}

	if err := git.SaveUseState(currentBranch, stashRef); err != nil {
		out.ErrorResult(err, "STATE_SAVE_ERROR")
		return err
	}

	if err := git.Checkout(targetWorktree.Branch); err != nil {
		git.ClearUseState()
		if stashRef != "" {
			git.StashPop()
		}
		out.ErrorResult(err, "CHECKOUT_ERROR")
		return err
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"branch":          targetWorktree.Branch,
			"previous_branch": currentBranch,
			"stashed":         stashRef != "",
		})
	}

	out.Success(fmt.Sprintf("Switched to branch: %s", targetWorktree.Branch))
	if stashRef != "" {
		out.Dim("  Changes stashed automatically")
	}
	out.Println()
	out.Info("Run 'lazywork worktree return' to go back")

	return nil
}

func runWorktreeReturn(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	if !git.IsMainWorktree() {
		err := fmt.Errorf("must be in main repository, not a worktree")
		out.ErrorResult(err, "NOT_MAIN_WORKTREE")
		return err
	}

	previousBranch, stashRef, err := git.LoadUseState()
	if err != nil {
		err := fmt.Errorf("no previous state found. Did you run 'worktree use' first?")
		out.ErrorResult(err, "NO_STATE")
		return err
	}

	if git.HasUncommittedChanges() {
		err := fmt.Errorf("you have uncommitted changes. Commit or stash them before returning")
		out.ErrorResult(err, "UNCOMMITTED_CHANGES")
		return err
	}

	if err := git.Checkout(previousBranch); err != nil {
		out.ErrorResult(err, "CHECKOUT_ERROR")
		return err
	}

	if stashRef != "" {
		if err := git.StashPop(); err != nil {
			out.Warning(fmt.Sprintf("Could not restore stash: %v", err))
		}
	}

	if err := git.ClearUseState(); err != nil {
		out.Warning(fmt.Sprintf("Could not clear state: %v", err))
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"branch":   previousBranch,
			"restored": stashRef != "",
		})
	}

	out.Success(fmt.Sprintf("Returned to branch: %s", previousBranch))
	if stashRef != "" {
		out.Dim("  Stashed changes restored")
	}

	return nil
}

func runWorktreeFinish(cmd *cobra.Command, args []string) error {
	out := output.New(jsonOutput, noColor)

	if !git.IsInsideWorkTree() {
		err := fmt.Errorf("not inside a git repository")
		out.ErrorResult(err, "NOT_GIT_REPO")
		return err
	}

	if !git.IsMainWorktree() {
		err := fmt.Errorf("must be in main repository, not a worktree")
		out.ErrorResult(err, "NOT_MAIN_WORKTREE")
		return err
	}

	currentBranch, err := git.CurrentBranch()
	if err != nil {
		out.ErrorResult(err, "BRANCH_ERROR")
		return err
	}

	mainBranch := git.GetMainBranch()
	if currentBranch != mainBranch {
		err := fmt.Errorf("must be on %s branch to finish a worktree", mainBranch)
		out.ErrorResult(err, "NOT_MAIN_BRANCH")
		return err
	}

	if git.HasUncommittedChanges() {
		err := fmt.Errorf("uncommitted changes detected. Commit or stash them first")
		out.ErrorResult(err, "UNCOMMITTED_CHANGES")
		return err
	}

	worktrees, err := git.ListWorktrees()
	if err != nil {
		out.ErrorResult(err, "WORKTREE_LIST_ERROR")
		return err
	}

	var secondaryWorktrees []git.Worktree
	for _, wt := range worktrees {
		if !wt.Bare && strings.Contains(wt.Path, string(filepath.Separator)+".worktrees"+string(filepath.Separator)) {
			secondaryWorktrees = append(secondaryWorktrees, wt)
		}
	}

	if len(secondaryWorktrees) == 0 {
		err := fmt.Errorf("no worktrees found")
		out.ErrorResult(err, "NO_WORKTREES")
		return err
	}

	var name string
	if len(args) > 0 {
		name = args[0]
	} else if out.IsTTY() {
		form := tui.WorktreeSelectForm(secondaryWorktrees, &name)
		if err := form.Run(); err != nil {
			return err
		}
	} else {
		err := fmt.Errorf("worktree name required")
		out.ErrorResult(err, "NAME_REQUIRED")
		return err
	}

	var targetWorktree *git.Worktree
	for _, wt := range secondaryWorktrees {
		if filepath.Base(wt.Path) == name || wt.Branch == name {
			targetWorktree = &wt
			break
		}
	}

	if targetWorktree == nil {
		err := fmt.Errorf("worktree '%s' not found", name)
		out.ErrorResult(err, "WORKTREE_NOT_FOUND")
		return err
	}

	if targetWorktree.Branch == "" {
		err := fmt.Errorf("worktree is in detached HEAD state, cannot merge")
		out.ErrorResult(err, "DETACHED_HEAD")
		return err
	}

	if err := git.Merge(targetWorktree.Branch); err != nil {
		out.Error(fmt.Sprintf("Merge failed: %v", err))
		out.Println()
		out.Info("Resolve conflicts and run 'git commit', then try again")
		return err
	}

	out.Success(fmt.Sprintf("Merged %s into %s", targetWorktree.Branch, mainBranch))

	var doCleanup bool
	if out.IsTTY() {
		form := tui.CleanupConfirmForm(filepath.Base(targetWorktree.Path), &doCleanup)
		if err := form.Run(); err != nil {
			return err
		}
	}

	if jsonOutput {
		return out.JSON(map[string]interface{}{
			"merged":  true,
			"branch":  targetWorktree.Branch,
			"cleanup": doCleanup,
		})
	}

	if doCleanup {
		if err := git.RemoveWorktree(targetWorktree.Path, false); err != nil {
			out.Warning(fmt.Sprintf("Could not remove worktree: %v", err))
		} else {
			out.Success(fmt.Sprintf("Removed worktree: %s", filepath.Base(targetWorktree.Path)))
		}

		if err := git.DeleteBranch(targetWorktree.Branch, false); err != nil {
			out.Warning(fmt.Sprintf("Could not delete branch: %v", err))
		} else {
			out.Success(fmt.Sprintf("Deleted branch: %s", targetWorktree.Branch))
		}
	}

	return nil
}
