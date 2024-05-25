package repo

import (
	"context"
	"fmt"
	"github.com/danesparza/package-assistant/internal/files"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"path/filepath"
	"time"
)

type gitRepoService struct {
	ProjectURL    string
	ProjectFolder string
	*git.Repository
}

type GitRepoService interface {
	Pull() error
	AddFile(srcFile string) error
	CommitAndPush() error
}

func NewGitRepoService(projectURL, projectFolder string, gitrepo *git.Repository) GitRepoService {
	return &gitRepoService{
		ProjectURL:    projectURL,
		ProjectFolder: projectFolder,
		Repository:    gitrepo,
	}
}

// InitPackageRepo makes sure that the package repo project folder is ready to use
// and that the git credential helper is set up and ready to use
func InitPackageRepo(ctx context.Context, projectUrl, projectFolder, username, password string) (*git.Repository, error) {
	log.Info().Msg("Initializing package repo...")
	_, err := os.Stat(projectFolder)
	if os.IsNotExist(err) {
		log.Info().Msg("project folder does not exist.  Git cloning ... ")
		_, err := git.PlainCloneContext(ctx, projectFolder, false, &git.CloneOptions{
			// The intended use of a GitHub personal access token is in replace of your password
			// because access tokens can easily be revoked.
			// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
			Auth: &http.BasicAuth{
				Username: username,
				Password: password,
			},
			URL:      projectUrl,
			Progress: os.Stdout,
		})

		if err != nil {
			log.Err(err).
				Str("projectFolder", projectFolder).
				Str("projectUrl", projectUrl).
				Msg("problem cloning repo")
			return nil, fmt.Errorf("problem cloning repo: %w", err)
		}
	}

	r, err := git.PlainOpen(projectFolder)
	if err != nil {
		log.Err(err).
			Str("projectFolder", projectFolder).
			Msg("problem opening repo")
		return nil, fmt.Errorf("problem opening repo: %w", err)
	}

	return r, nil
}

// Pull pulls (syncs) upstream changes into the local repo
func (g gitRepoService) Pull() error {
	// Get the working directory for the repository
	w, err := g.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("problem getting working tree when pulling: %w", err)
	}

	//	Pull
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		return fmt.Errorf("problem pulling repository: %w", err)
	}

	return nil
}

// AddFile add the file to the repository
func (g gitRepoService) AddFile(srcFile string) error {
	//	Get just the filename we're trying to process:
	_, currentFileName := filepath.Split(srcFile)

	err := files.Copy(srcFile, path.Join(g.ProjectFolder, currentFileName), os.ModePerm)
	if err != nil {
		return fmt.Errorf("problem copying file: %w", err)
	}

	// Get the working directory for the repository
	w, err := g.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("problem getting working tree when adding: %w", err)
	}

	//	Add the file
	_, err = w.Add(srcFile)
	if err != nil {
		return fmt.Errorf("problem adding the file: %w", err)
	}

	return nil
}

// CommitAndPush commits the changes and pushes to the remote
func (g gitRepoService) CommitAndPush() error {
	// Get the working directory for the repository
	w, err := g.Repository.Worktree()
	if err != nil {
		return fmt.Errorf("problem getting working tree when committing: %w", err)
	}

	//	Commit the file(s)
	_, err = w.Commit("package repo bot commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("problem committing: %w", err)
	}

	//	Push
	err = g.Repository.Push(&git.PushOptions{})
	if err != nil {
		return fmt.Errorf("problem pushing: %w", err)
	}

	return nil
}
