package api

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"path"
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

	MAX_UPLOAD_SIZE := viper.GetInt64("upload.bytelimit")
	UploadPath := viper.GetString("upload.path")

	//	First check for maximum uplooad size and return an error if we exceed it.
	req.Body = http.MaxBytesReader(rw, req.Body, MAX_UPLOAD_SIZE)
	if err := req.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		err = fmt.Errorf("uploaded file is too big: %w", err)
		sendErrorResponse(rw, err, http.StatusRequestEntityTooLarge)
		return
	}

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, fileHeader, err := req.FormFile("file")
	if err != nil {
		err = fmt.Errorf("error retrieving file from the 'file' form element: %w", err)
		sendErrorResponse(rw, err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create the uploads folder if it doesn't
	// already exist
	err = os.MkdirAll(UploadPath, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("error creating uploads path: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	// Create a new file in the uploads directory
	destinationFile := path.Join(UploadPath, fileHeader.Filename)
	dst, err := os.Create(destinationFile)
	if err != nil {
		err = fmt.Errorf("error creating file: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		err = fmt.Errorf("error saving file: %w", err)
		sendErrorResponse(rw, err, http.StatusInternalServerError)
		return
	}

	//	Process the file
	//	Does the git repo exist? If not, clone it (it's big!):
	//  cd /data
	//  git clone ${PACKASSIST_GITHUB_PROJECTURL}

	//	Switch to the project directory:
	//	cd /data/package-assistant

	//  Use git config credential.helper:
	//  git config credential.helper '!f() { sleep 1; echo "username=${PACKASSIST_GITHUB_USER}"; echo "password=${PACKASSIST_GITHUB_PASSWORD}"; }; f'

	//	Make sure to set an identity:
	//	git config --global user.email "danesparza@cagedtornado.com"
	//	git config --global user.name "Package repo bot"
	//	Now you should be able to git add . / git commit -m "Some message"

	//	If we've gotten this far, indicate a successful upload
	response := SystemResponse{
		Message: fmt.Sprintf("File uploaded: %v", fileHeader.Filename),
	}

	//	Serialize to JSON & return the response:
	rw.WriteHeader(http.StatusCreated)
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(rw).Encode(response)
}
