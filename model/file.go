package model

//File is the main data struct of our file system
type File struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
	Size int64  `json:"Size"`
}

type FileList struct {
	Files []File
}
