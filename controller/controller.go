package controller

import (
	"html/template"
	"log"
	"net/http"
)

const indexPage = "public/index.html"

var (
	fileController file
	homeTemplate   *template.Template
)

//Startup - This is the startup method
func Startup(templates map[string]*template.Template) {

	fileController.uploadTemplate = templates["upload.html"]
	fileController.listTemplate = templates["list.html"]
	fileController.uploadComplete = templates["complete.html"]

	homeTemplate = templates["home.html"]

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/favicon.ico", faviconHandler)
	fileController.registerRoutes()

	http.Handle("/img/", http.FileServer(http.Dir("public")))
	http.Handle("/css/", http.FileServer(http.Dir("public")))
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/favicon.ico")
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
	w.Header().Add("Content-Type", "text/html")
	homeTemplate.Execute(w, nil)
}
