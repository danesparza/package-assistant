package git

import (
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
func InitPackageRepo(projectUrl, baseFolder, projectFolder string) error {
	//	Does the git repo exist? If not, clone it (it's big!):
	_, err := os.Stat(projectFolder)
	if os.IsNotExist(err) {
		//  cd /data (baseFolder)
		//  git clone ${PACKASSIST_GITHUB_PROJECTURL}
	}

	//	Switch to the project directory (projectFolder):
	//	cd /data/package-assistant

	//  Use git config credential.helper:
	//  git config credential.helper '!f() { sleep 1; echo "username=${PACKASSIST_GITHUB_USER}"; echo "password=${PACKASSIST_GITHUB_PASSWORD}"; }; f'

	//	Make sure to set an identity:
	//	git config --global user.email "danesparza@cagedtornado.com"
	//	git config --global user.name "Package repo bot"
	//	Now you should be able to git add . / git commit -m "Some message" / git push
	return nil
}
