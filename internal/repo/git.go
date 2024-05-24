package repo

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/rs/zerolog/log"
	"os"
)

type gitRepoService struct {
	ProjectURL    string
	ProjectFolder string
}

func (g gitRepoService) Pull() error {
	//TODO implement me
	panic("implement me")
}

func (g gitRepoService) AddFile(srcFile string) error {
	//TODO implement me
	panic("implement me")
}

func (g gitRepoService) CommitAndPush() error {
	//TODO implement me
	panic("implement me")
}

type GitRepoService interface {
	Pull() error
	AddFile(srcFile string) error
	CommitAndPush() error
}

func NewGitRepoService(projectURL, projectFolder string) GitRepoService {
	return &gitRepoService{
		ProjectURL:    projectURL,
		ProjectFolder: projectFolder,
	}
}

// InitPackageRepo makes sure that the package repo project folder is ready to use
// and that the git credential helper is set up and ready to use
func InitPackageRepo(ctx context.Context, projectUrl, baseFolder, projectFolder, username, password string) (*git.Repository, error) {
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
