package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Repository struct {
	path string
}

func NewRepository(path string) *Repository {
	return &Repository{path: path}
}

func (r *Repository) CurrentBranch() (string, error) {
	output, err := r.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func (r *Repository) DefaultBranch() (string, error) {
	output, err := r.runGitCommand("symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		output, err = r.runGitCommand("rev-parse", "--abbrev-ref", "origin/HEAD")
		if err != nil {
			branches := []string{"main", "master"}
			for _, branch := range branches {
				if _, err := r.runGitCommand("show-ref", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
					return branch, nil
				}
			}
			
			output, err = r.runGitCommand("rev-parse", "--abbrev-ref", "HEAD")
			if err == nil {
				return strings.TrimSpace(output), nil
			}
			
			return "", fmt.Errorf("failed to determine default branch")
		}
	}
	
	branch := strings.TrimPrefix(strings.TrimSpace(output), "refs/remotes/origin/")
	branch = strings.TrimPrefix(branch, "origin/")
	return branch, nil
}

func (r *Repository) DiffAgainstDefault() (string, error) {
	defaultBranch, err := r.DefaultBranch()
	if err != nil {
		return "", err
	}
	
	var mergeBase string
	mergeBase, err = r.runGitCommand("merge-base", "HEAD", "origin/"+defaultBranch)
	if err != nil {
		mergeBase, err = r.runGitCommand("merge-base", "HEAD", defaultBranch)
		if err != nil {
			return "", fmt.Errorf("failed to find merge base: %w", err)
		}
	}
	mergeBase = strings.TrimSpace(mergeBase)
	
	diff, err := r.runGitCommand("diff", mergeBase+"...HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to generate diff: %w", err)
	}
	
	return diff, nil
}

func (r *Repository) GetRemoteURL() (string, error) {
	output, err := r.runGitCommand("remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func (r *Repository) GetChangedFiles() ([]string, error) {
	defaultBranch, err := r.DefaultBranch()
	if err != nil {
		return nil, err
	}
	
	var mergeBase string
	mergeBase, err = r.runGitCommand("merge-base", "HEAD", "origin/"+defaultBranch)
	if err != nil {
		mergeBase, err = r.runGitCommand("merge-base", "HEAD", defaultBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to find merge base: %w", err)
		}
	}
	mergeBase = strings.TrimSpace(mergeBase)
	
	output, err := r.runGitCommand("diff", "--name-only", mergeBase+"...HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	
	return files, nil
}

func (r *Repository) PushCurrentBranch() error {
	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	
	_, err = r.runGitCommand("push", "-u", "origin", currentBranch)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}
	
	return nil
}

func (r *Repository) runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if r.path != "" {
		cmd.Dir = r.path
	}
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w, stderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}