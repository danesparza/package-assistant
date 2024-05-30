package api

import (
	"encoding/json"
	"fmt"
	"github.com/danesparza/package-assistant/internal/debian"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// UploadPackage godoc
// @Summary Upload a package
// @Description Upload package
// @Tags package
// @Accept  mpfd
// @Produce  json
// @Param file formData file true "The file to upload"
// @Success 200 {object} api.SystemResponse
// @Failure 400 {object} api.ErrorResponse
// @Failure 413 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /package [post]
func (service Service) UploadPackage(rw http.ResponseWriter, req *http.Request) {

	//	Get configs
	MAX_UPLOAD_SIZE := viper.GetInt64("upload.bytelimit")
	UploadPath := viper.GetString("upload.path")
	RepoPath := viper.GetString("github.projectfolder")
	gpgPassword := viper.GetString("gpg.password")
	gitName := viper.GetString("git.name")
	gitEmail := viper.GetString("git.email")
	githubUser := viper.GetString("github.user")
	githubPassword := viper.GetString("github.password")

	//	First check the auth token and make sure it exists on the header:
	log.Debug().Msg("Validating X-PackAuth header")
	authToken := req.Header.Get("X-PackAuth")
	if strings.TrimSpace(authToken) != strings.TrimSpace(viper.GetString("auth.token")) {
		err := fmt.Errorf("X-PackAuth token invalid")
		sendErrorResponse(rw, err, http.StatusUnauthorized)
		return
	}

	//	First check for maximum uplooad size and return an error if we exceed it.
	log.Debug().Int64("MAX_UPLOAD_SIZE", MAX_UPLOAD_SIZE).Msg("Checking size vs max upload size")
	req.Body = http.MaxBytesReader(rw, req.Body, MAX_UPLOAD_SIZE)
	if err := req.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		err = fmt.Errorf("uploaded file is too big: %w", err)
		sendErrorResponse(rw, err, http.StatusRequestEntityTooLarge)
		return
	}

	// FormFile returns the first file for the given key `file`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	log.Debug().Msg("Checking for 'file' form element and fetching")
	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		err = fmt.Errorf("error retrieving file from the 'file' form element: %w", err)
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the uploads folder if it doesn't
	// already exist
	log.Debug().Str("UploadPath", UploadPath).Msg("Creating upload path if it doesn't exist")
	err = os.MkdirAll(UploadPath, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("error creating uploads path: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	destinationFile := path.Join(UploadPath, fileHeader.Filename)
	log.Debug().Str("destination file", destinationFile).Msg("Creating file in uploads directory")
	dst, err := os.Create(destinationFile)
	if err != nil {
		err = fmt.Errorf("error creating file: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	log.Debug().Msg("Copying file data to the destination file")
	_, err = io.Copy(dst, file)
	if err != nil {
		err = fmt.Errorf("error saving file: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Process the file

	//	ci-pre.sh (switch to repo folder and git pull)
	log.Debug().Msg("Performing a repo pull")
	err = service.RepoSvc.Pull()
	if err != nil {
		err = fmt.Errorf("error refreshing repo: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Move file to repo folder
	repoFile := path.Join(RepoPath, fileHeader.Filename)
	log.Debug().Str("repoFile", repoFile).Msg("Moving file to the repo path")
	err = os.Rename(destinationFile, repoFile)
	if err != nil {
		err = fmt.Errorf("error moving file to repo: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//  ci-refresh.sh / refresh-packages.sh (Perform dpkg-scanpackages, gzip and sign using gpg)
	log.Debug().Msg("Refreshing packages")
	err = debian.RefreshPackages(req.Context(), gpgPassword, gitEmail, RepoPath)
	if err != nil {
		err = fmt.Errorf("error refreshing packages: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	ci-post.sh (git add / git commit / git push)
	log.Debug().Msg("Adding all changes and preparing to commit")
	err = service.RepoSvc.AddAll()
	if err != nil {
		err = fmt.Errorf("error adding changes in repo: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	log.Debug().Msg("Committing and pushing changes")
	err = service.RepoSvc.CommitAndPush(githubUser, githubPassword, gitName, gitEmail)
	if err != nil {
		err = fmt.Errorf("error committing and pushing: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	If we've gotten this far, indicate a successful upload
	response := SystemResponse{
		Message: fmt.Sprintf("File uploaded: %v", fileHeader.Filename),
	}

	//	Serialize to JSON & return the response:
	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
	log.Debug().Msg("Complete!")
}
