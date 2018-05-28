package model

import (
	"errors"
	"io"
	"regexp"

	uuid "github.com/satori/go.uuid"
)

//File is the main data struct of our file system
type File struct {
	ID          string        `json:"ID"`
	Name        string        `json:"Name"`
	Size        int64         `json:"Size"`
	Blob        io.ReadCloser `json:"-"`
	ContentType string        `json:"-"`
	AccessToken string        `json:"-"` //Future use
	UploadedBy  string        `json:"-"` //Future use
}

//FileResponse - server repsonse to an uploaded file - returns file id
type FileResponse struct {
	ID          string `json:"ID"`
	AccessToken string `json:"-"` //Future use
}

//FileList - list of files
type FileList struct {
	Files []File `json:"Files"`
}

//GetFiles - Returns list of files
func GetFiles() ([]File, error) {
	return fileRepo.ListObjects()
}

//UploadFileToRepo - Uploads file to repository
func UploadFileToRepo(file []byte, fileName string) (string, string, error) {
	if len(fileName) == 0 {
		return "", "", errors.New("File name is required")
	}

	u1 := uuid.Must(uuid.NewV4()).String()
	key := u1

	accessToken, err := fileRepo.PutObject(key, file, fileName)
	return accessToken, key, err
}

//GetFileFromRepo - Retrieves file from repository
func GetFileFromRepo(key string) (File, error) {
	if len(key) == 0 || !isValidUUID(key) {
		return File{}, errors.New("Empty or invalid key")
	}
	return fileRepo.GetObject(key)
}

func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}
