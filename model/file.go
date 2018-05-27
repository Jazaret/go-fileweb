package model

import (
	"io"

	uuid "github.com/satori/go.uuid"
)

//File is the main data struct of our file system
type File struct {
	ID          string        `json:"ID"`
	Name        string        `json:"Name"`
	Size        int64         `json:"Size"`
	Blob        io.ReadCloser `json:"Blob"`
	ContentType string        `json:"ContentType`
}

//FileResponse - server repsonse to an uploaded file - returns file id
type FileResponse struct {
	ID string `json:"ID"`
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
func UploadFileToRepo(file []byte, fileName string) (string, error) {
	u1 := uuid.Must(uuid.NewV4()).String()
	key := u1

	err := fileRepo.PutObject(key, file, fileName)
	return key, err
}

//GetFileFromRepo - Retrieves file from repository
func GetFileFromRepo(key string) (File, error) {
	return fileRepo.GetObject(key)
}
