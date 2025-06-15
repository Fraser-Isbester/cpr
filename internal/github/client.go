package github

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	ctx    context.Context
}

func NewClient(token string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

func (c *Client) CreateOrUpdatePullRequest(owner, repo, title, body, head, base string, draft bool) (*github.PullRequest, bool, error) {
	// First, check if a PR already exists for this branch
	existingPR, err := c.GetPullRequestForBranch(owner, repo, head)
	if err != nil {
		return nil, false, err
	}

	if existingPR != nil {
		// Update existing PR
		updatedPR, err := c.UpdatePullRequest(owner, repo, existingPR.GetNumber(), title, body)
		if err != nil {
			return nil, false, fmt.Errorf("failed to update pull request: %w", err)
		}
		return updatedPR, true, nil
	}

	// Create new PR
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
		Draft: github.Bool(draft),
	}

	pullRequest, _, err := c.client.PullRequests.Create(c.ctx, owner, repo, pr)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pullRequest, false, nil
}

func (c *Client) GetPullRequestForBranch(owner, repo, branch string) (*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		Head:  owner + ":" + branch,
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	pulls, _, err := c.client.PullRequests.List(c.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}

	if len(pulls) > 0 {
		return pulls[0], nil
	}

	return nil, nil
}

func (c *Client) UpdatePullRequest(owner, repo string, number int, title, body string) (*github.PullRequest, error) {
	update := &github.PullRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}

	pr, _, err := c.client.PullRequests.Edit(c.ctx, owner, repo, number, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update pull request: %w", err)
	}

	return pr, nil
}

func (c *Client) GetPullRequestTemplate(owner, repo string) (string, error) {
	// Common PR template locations
	templatePaths := []string{
		".github/pull_request_template.md",
		".github/PULL_REQUEST_TEMPLATE.md",
		"pull_request_template.md",
		"PULL_REQUEST_TEMPLATE.md",
		".github/PULL_REQUEST_TEMPLATE/pull_request_template.md",
		"docs/pull_request_template.md",
		"docs/PULL_REQUEST_TEMPLATE.md",
	}

	for _, path := range templatePaths {
		content, _, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, path, &github.RepositoryContentGetOptions{})
		if err == nil && content != nil {
			decoded, err := content.GetContent()
			if err == nil {
				return decoded, nil
			}
		}
	}

	// Check for multiple templates in .github/PULL_REQUEST_TEMPLATE/
	_, dirs, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, ".github/PULL_REQUEST_TEMPLATE", &github.RepositoryContentGetOptions{})
	if err == nil && len(dirs) > 0 {
		// If multiple templates exist, use the first one
		content, _, _, err := c.client.Repositories.GetContents(c.ctx, owner, repo, dirs[0].GetPath(), &github.RepositoryContentGetOptions{})
		if err == nil && content != nil {
			decoded, err := content.GetContent()
			if err == nil {
				return decoded, nil
			}
		}
	}

	return "", nil
}

func (c *Client) CreatePullRequest(owner, repo, title, body, head, base string, draft bool) (*github.PullRequest, error) {
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
		Draft: github.Bool(draft),
	}

	pullRequest, _, err := c.client.PullRequests.Create(c.ctx, owner, repo, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return pullRequest, nil
}

func GetToken() (string, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return token, nil
	}

	token = os.Getenv("GH_TOKEN")
	if token != "" {
		return token, nil
	}

	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not authenticated. Please run 'gh auth login' or set GITHUB_TOKEN environment variable")
	}

	cmd = exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get token from gh CLI: %w", err)
	}

	token = strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("gh CLI returned empty token")
	}

	return token, nil
}

func ParseGitRemoteURL(remoteURL string) (owner string, repo string, err error) {
	remoteURL = strings.TrimSpace(remoteURL)

	if strings.HasPrefix(remoteURL, "git@github.com:") {
		path := strings.TrimPrefix(remoteURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH remote URL format")
		}
		return parts[0], parts[1], nil
	}

	if strings.HasPrefix(remoteURL, "https://") || strings.HasPrefix(remoteURL, "http://") {
		u, err := url.Parse(remoteURL)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse remote URL: %w", err)
		}

		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid HTTPS remote URL format")
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unsupported remote URL format")
}

// ApplyTemplate fills in a PR template with the provided content
func ApplyTemplate(template, title, summary string) string {
	// If no template, just return the summary
	if template == "" {
		return summary
	}

	// Replace common template placeholders
	result := template
	
	// Replace title placeholders
	result = strings.ReplaceAll(result, "{{title}}", title)
	result = strings.ReplaceAll(result, "{{TITLE}}", title)
	result = strings.ReplaceAll(result, "[Title]", title)
	result = strings.ReplaceAll(result, "[TITLE]", title)
	
	// Replace description/summary placeholders
	result = strings.ReplaceAll(result, "{{description}}", summary)
	result = strings.ReplaceAll(result, "{{DESCRIPTION}}", summary)
	result = strings.ReplaceAll(result, "{{summary}}", summary)
	result = strings.ReplaceAll(result, "{{SUMMARY}}", summary)
	result = strings.ReplaceAll(result, "[Description]", summary)
	result = strings.ReplaceAll(result, "[DESCRIPTION]", summary)
	result = strings.ReplaceAll(result, "[Summary]", summary)
	result = strings.ReplaceAll(result, "[SUMMARY]", summary)

	// Look for sections to fill in
	// Common patterns: ## Summary, ## Description, ## What, ## Why, ## Changes
	patterns := []string{
		"## Summary",
		"## Description", 
		"## What",
		"## Changes",
		"### Summary",
		"### Description",
		"### What",
		"### Changes",
	}

	for _, pattern := range patterns {
		idx := strings.Index(result, pattern)
		if idx >= 0 {
			// Find the next section header or end of string
			endIdx := len(result)
			for _, p := range []string{"##", "###", "---"} {
				if next := strings.Index(result[idx+len(pattern):], p); next >= 0 {
					endIdx = idx + len(pattern) + next
					break
				}
			}
			
			// Insert the summary after the header
			before := result[:idx+len(pattern)]
			after := result[endIdx:]
			result = before + "\n\n" + summary + "\n" + after
			break // Only fill the first matching section
		}
	}

	return result
}