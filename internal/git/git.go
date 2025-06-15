package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type Repository struct {
	repo *git.Repository
	path string
}

func NewRepository(path string) *Repository {
	if path == "" {
		path = "."
	}
	return &Repository{path: path}
}

func (r *Repository) open() error {
	if r.repo != nil {
		return nil
	}
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	r.repo = repo
	return nil
}

func (r *Repository) CurrentBranch() (string, error) {
	if err := r.open(); err != nil {
		return "", err
	}

	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	return "HEAD", nil
}

func (r *Repository) DefaultBranch() (string, error) {
	if err := r.open(); err != nil {
		return "", err
	}

	// Try to get the default branch from origin
	remote, err := r.repo.Remote("origin")
	if err != nil {
		// If no remote, check common default branches
		return r.findLocalDefaultBranch()
	}

	// Get remote refs
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return r.findLocalDefaultBranch()
	}

	// Look for HEAD reference
	for _, ref := range refs {
		if ref.Name().String() == "HEAD" && ref.Target() != "" {
			target := ref.Target().String()
			if strings.HasPrefix(target, "refs/heads/") {
				return strings.TrimPrefix(target, "refs/heads/"), nil
			}
		}
	}

	// Check for symbolic-ref
	for _, ref := range refs {
		if ref.Name().String() == "refs/remotes/origin/HEAD" {
			target := ref.Target().String()
			if strings.HasPrefix(target, "refs/remotes/origin/") {
				return strings.TrimPrefix(target, "refs/remotes/origin/"), nil
			}
		}
	}

	return r.findLocalDefaultBranch()
}

func (r *Repository) findLocalDefaultBranch() (string, error) {
	// Check common default branch names
	branches := []string{"main", "master"}
	for _, branch := range branches {
		ref, err := r.repo.Reference(plumbing.NewBranchReferenceName(branch), false)
		if err == nil && ref != nil {
			return branch, nil
		}
	}

	// If no common branches found, return current branch
	current, err := r.CurrentBranch()
	if err != nil {
		return "", fmt.Errorf("failed to determine default branch")
	}
	return current, nil
}

func (r *Repository) DiffAgainstDefault() (string, error) {
	if err := r.open(); err != nil {
		return "", err
	}

	defaultBranch, err := r.DefaultBranch()
	if err != nil {
		return "", err
	}

	// Get HEAD commit
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	headCommit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Find merge base
	var baseCommit *object.Commit
	
	// Try origin/defaultBranch first
	originRef, err := r.repo.Reference(plumbing.NewRemoteReferenceName("origin", defaultBranch), true)
	if err == nil {
		baseCommit, err = r.findMergeBase(headCommit, originRef.Hash())
	}
	
	// If that fails, try local defaultBranch
	if err != nil || baseCommit == nil {
		localRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(defaultBranch), true)
		if err != nil {
			return "", fmt.Errorf("failed to find reference for %s: %w", defaultBranch, err)
		}
		baseCommit, err = r.findMergeBase(headCommit, localRef.Hash())
		if err != nil {
			return "", fmt.Errorf("failed to find merge base: %w", err)
		}
	}

	// Generate diff between merge base and HEAD
	baseTree, err := baseCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get head tree: %w", err)
	}

	patch, err := baseTree.Patch(headTree)
	if err != nil {
		return "", fmt.Errorf("failed to generate patch: %w", err)
	}

	return patch.String(), nil
}

func (r *Repository) findMergeBase(commit *object.Commit, targetHash plumbing.Hash) (*object.Commit, error) {
	targetCommit, err := r.repo.CommitObject(targetHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get target commit: %w", err)
	}

	// Find common ancestor using merge base algorithm
	commitAncestors := make(map[plumbing.Hash]bool)
	
	// Walk ancestors of commit
	err = object.NewCommitIterCTime(commit, nil, nil).ForEach(func(c *object.Commit) error {
		commitAncestors[c.Hash] = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Walk ancestors of target until we find common one
	var mergeBase *object.Commit
	err = object.NewCommitIterCTime(targetCommit, nil, nil).ForEach(func(c *object.Commit) error {
		if commitAncestors[c.Hash] {
			mergeBase = c
			return fmt.Errorf("found") // Stop iteration
		}
		return nil
	})
	
	if err != nil && err.Error() != "found" {
		return nil, err
	}

	if mergeBase == nil {
		return nil, fmt.Errorf("no common ancestor found")
	}

	return mergeBase, nil
}

func (r *Repository) GetRemoteURL() (string, error) {
	if err := r.open(); err != nil {
		return "", err
	}

	remote, err := r.repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("no URLs configured for remote origin")
	}

	return urls[0], nil
}

func (r *Repository) GetChangedFiles() ([]string, error) {
	if err := r.open(); err != nil {
		return nil, err
	}

	defaultBranch, err := r.DefaultBranch()
	if err != nil {
		return nil, err
	}

	// Get HEAD commit
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	headCommit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Find merge base
	var baseCommit *object.Commit
	
	// Try origin/defaultBranch first
	originRef, err := r.repo.Reference(plumbing.NewRemoteReferenceName("origin", defaultBranch), true)
	if err == nil {
		baseCommit, err = r.findMergeBase(headCommit, originRef.Hash())
	}
	
	// If that fails, try local defaultBranch
	if err != nil || baseCommit == nil {
		localRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(defaultBranch), true)
		if err != nil {
			return nil, fmt.Errorf("failed to find reference for %s: %w", defaultBranch, err)
		}
		baseCommit, err = r.findMergeBase(headCommit, localRef.Hash())
		if err != nil {
			return nil, fmt.Errorf("failed to find merge base: %w", err)
		}
	}

	// Get trees
	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get head tree: %w", err)
	}

	// Get changes between trees
	changes, err := object.DiffTree(baseTree, headTree)
	if err != nil {
		return nil, fmt.Errorf("failed to diff trees: %w", err)
	}

	// Extract file names
	var files []string
	for _, change := range changes {
		name := change.From.Name
		if name == "" {
			name = change.To.Name
		}
		if name != "" {
			files = append(files, name)
		}
	}

	return files, nil
}

func (r *Repository) PushCurrentBranch() error {
	if err := r.open(); err != nil {
		return err
	}

	currentBranch, err := r.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get authentication
	auth, err := r.getAuth()
	if err != nil {
		return fmt.Errorf("failed to get authentication: %w", err)
	}

	// Push to origin
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", currentBranch, currentBranch))
	err = r.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{refSpec},
		Auth:       auth,
		Progress:   nil,
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push branch: %w", err)
	}

	return nil
}

func (r *Repository) getAuth() (transport.AuthMethod, error) {
	// For now, rely on system git configuration (SSH keys, etc.)
	// In a production environment, you might want to support various auth methods
	return nil, nil
}