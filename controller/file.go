package controller

import (
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jazaret/go-fileweb/model"
)

type file struct {
	uploadTemplate *template.Template
	listTemplate   *template.Template
	uploadComplete *template.Template
}

func (f file) registerRoutes() {
	http.HandleFunc("/api/list", f.listFilesAPI)
	http.HandleFunc("/api/download/", f.downloadFileToClient)
	http.HandleFunc("/api/upload", f.receiveFileFromClient)

	http.HandleFunc("/upload", f.receiveFileFromClient)
	http.HandleFunc("/download/", f.downloadFileToClient)
	http.HandleFunc("/list", f.listFiles)
}

func (f file) receiveFileFromClient(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		log.Printf("HTTP POST received on upload...\n")

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error on FormFile %s\n", err)
			w.Write([]byte(err.Error()))
			return
		}
		defer file.Close()

		if header.Filename == "" {
			log.Printf("File does not exist %s\n", err)
			w.Write([]byte("File does not exist"))
			return
		}

		buffer, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("Error on ReadAll %s\n", err)
			w.Write([]byte(err.Error()))
			return
		}

		result, err2 := model.UploadFileToRepo(buffer, header.Filename)
		if err2 != nil {
			log.Printf("Error on Upload %s\n", err)
			w.Write([]byte(err.Error()))
			return
		}

		data := model.FileResponse{ID: result}
		if strings.Contains(r.URL.Path, "/api/") {
			jData, err3 := json.Marshal(data)
			if err3 != nil {
				log.Printf("Error on Marshal %s\n", err)
				w.Write([]byte(err.Error()))
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.Write(jData)
		} else {
			w.Header().Add("Content-Type", "text/html")
			f.uploadComplete.Execute(w, data)
		}
	} else {
		if strings.Contains(r.URL.Path, "/api/") {
			w.Write([]byte("API - Please send POST of FormFile with attribute name of 'file' or use website url"))
		} else {
			w.Header().Add("Content-Type", "text/html")
			f.uploadTemplate.Execute(w, nil)
		}
	}
}

func (f file) listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := model.GetFiles()
	if err != nil {
		log.Printf("Failed to GetList: %s\n", err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	fileList := model.FileList{
		Files: files,
	}
	w.Header().Add("Content-Type", "text/html")
	f.listTemplate.Execute(w, fileList)
}

func (f file) listFilesAPI(w http.ResponseWriter, r *http.Request) {
	files, err := model.GetFiles()
	if err != nil {
		log.Printf("Failed to GetList: %s\n", err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	jData, err2 := json.Marshal(files)
	if err2 != nil {
		log.Printf("Error on Marshal %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(jData)
}

func (f file) downloadFileToClient(w http.ResponseWriter, r *http.Request) {
	keySet := strings.Split(r.URL.Path, "download/")

	if len(keySet) < 1 {
		log.Printf("Error - key not specified\n")
		w.Write([]byte("Error - key not specified"))
		return
	}

	key := keySet[1]

	result, err := model.GetFileFromRepo(key)

	if err != nil {
		log.Printf("Error on GetObject %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}

	fileName := result.Name

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(result.Size, 10))

	io.Copy(w, result.Blob)

	defer result.Blob.Close()
}
