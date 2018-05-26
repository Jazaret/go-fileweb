package model

//File is the main data struct of our file system
type File struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
	Size int64  `json:"Size"`
}

type FileResponse struct {
	ID string `json:"ID"`
}

type FileList struct {
	Files []File
}
