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
