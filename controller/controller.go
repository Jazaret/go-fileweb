package controller

import (
	"html/template"
	"log"
	"net/http"
)

const indexPage = "public/index.html"

var (
	fileController file
)

//Startup - This is the startup method
func Startup(templates map[string]*template.Template) {
	fileController.init()

	fileController.uploadTemplate = templates["upload.html"]
	fileController.listTemplate = templates["list.html"]
	fileController.uploadComplete = templates["complete.html"]

	http.HandleFunc("/", handleRoot)
	fileController.registerRoutes()

	http.Handle("/img/", http.FileServer(http.Dir("public")))
	http.Handle("/css/", http.FileServer(http.Dir("public")))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
	http.ServeFile(w, r, indexPage)
}
