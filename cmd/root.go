package cmd

import (
	"fmt"
	"os"

	"github.com/fraser-isbester/cpr/internal/commit"
	"github.com/fraser-isbester/cpr/internal/git"
	"github.com/fraser-isbester/cpr/internal/github"
	"github.com/spf13/cobra"
)

var (
	title   string
	body    string
	draft   bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "cpr",
	Short: "Create a GitHub pull request with auto-generated Angular format title",
	Long: `cpr is a lightweight CLI tool that creates GitHub pull requests with
automatically generated Angular format titles and summaries based on your git diff.

It analyzes the changes between your current branch and the default branch,
generates an appropriate PR title (e.g., "feat: add new feature", "fix: resolve bug"),
and creates a comprehensive PR summary.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := createPR(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&title, "title", "t", "", "Custom PR title (overrides auto-generation)")
	rootCmd.Flags().StringVarP(&body, "body", "b", "", "Custom PR body (overrides auto-generation)")
	rootCmd.Flags().BoolVarP(&draft, "draft", "d", false, "Create PR as draft")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}

func createPR() error {
	repo := git.NewRepository("")

	currentBranch, err := repo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if currentBranch == "HEAD" {
		return fmt.Errorf("in detached HEAD state, please checkout a branch")
	}

	defaultBranch, err := repo.DefaultBranch()
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	if currentBranch == defaultBranch {
		return fmt.Errorf("cannot create PR from default branch '%s'", defaultBranch)
	}

	if verbose {
		fmt.Printf("Current branch: %s\n", currentBranch)
		fmt.Printf("Default branch: %s\n", defaultBranch)
	}

	diff, err := repo.DiffAgainstDefault()
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if diff == "" {
		if verbose {
			fmt.Printf("No changes detected between %s and %s\n", currentBranch, defaultBranch)
		}
		return fmt.Errorf("no changes detected between %s and %s", currentBranch, defaultBranch)
	}

	changedFiles, err := repo.GetChangedFiles()
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	analyzer := commit.NewAnalyzer(diff, changedFiles)

	if title == "" {
		title = analyzer.GenerateTitle()
		if verbose {
			fmt.Printf("Generated title: %s\n", title)
		}
	}

	if body == "" {
		body = analyzer.GenerateSummary()
		if verbose {
			fmt.Printf("Generated body:\n%s\n", body)
		}
	}

	remoteURL, err := repo.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	owner, repoName, err := github.ParseGitRemoteURL(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to parse remote URL: %w", err)
	}

	if verbose {
		fmt.Printf("Repository: %s/%s\n", owner, repoName)
	}

	// Push the current branch to origin if needed
	if err := repo.PushCurrentBranch(); err != nil {
		if verbose {
			fmt.Printf("Note: %v\n", err)
		}
	}

	token, err := github.GetToken()
	if err != nil {
		return err
	}

	client := github.NewClient(token)

	// Check for PR template
	template, err := client.GetPullRequestTemplate(owner, repoName)
	if err != nil && verbose {
		fmt.Printf("Failed to fetch PR template: %v\n", err)
	}

	// Apply template if found
	if template != "" {
		if verbose {
			fmt.Printf("Found PR template, applying...\n")
		}
		body = github.ApplyTemplate(template, title, body)
	}

	// Create or update PR
	pr, updated, err := client.CreateOrUpdatePullRequest(owner, repoName, title, body, currentBranch, defaultBranch, draft)
	if err != nil {
		return fmt.Errorf("failed to create/update pull request: %w", err)
	}

	if updated {
		fmt.Printf("Pull request updated: %s\n", pr.GetHTMLURL())
	} else {
		fmt.Printf("Pull request created: %s\n", pr.GetHTMLURL())
	}

	return nil
}