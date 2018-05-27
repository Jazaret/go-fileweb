package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/Jazaret/go-fileweb/controller"
	"github.com/Jazaret/go-fileweb/model"
)

var (
	region     = "us-east-1"        //os.Getenv("AWS_REGION")
	bucketName = "jazar-testbucket" //os.Getenv("BUCKET_NAME")
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	f, _ := os.Create("/var/log/golang/golang-server.log")
	defer f.Close()
	log.SetOutput(f)

	model.InitRepository(region, bucketName)

	templates := populateTemplates()
	controller.Startup(templates)

	log.Printf("Listening on port %s\n\n", port)
	http.ListenAndServe(":"+port, nil)
}

//Create template for our website
func populateTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template)
	const basePath = "templates"
	layout := template.Must(template.ParseFiles(basePath + "/_layout.html"))

	//Content Templates
	dir, err := os.Open(basePath + "/content")
	if err != nil {
		panic("Failed to read contents of blocks directiory: " + err.Error())
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		panic("Failed to read contents of content directiory: " + err.Error())
	}
	for _, fi := range fis {
		f, err := os.Open(basePath + "/content/" + fi.Name())
		if err != nil {
			panic("Failed to open template '" + fi.Name() + "'")
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			panic("Failed to read content from file '" + fi.Name() + "'")
		}
		f.Close()

		tmpl := template.Must(layout.Clone())
		_, err = tmpl.Parse(string(content))
		if err != nil {
			panic("Failed to parse contents of '" + fi.Name() + "' as template. Error - " + err.Error())
		}
		result[fi.Name()] = tmpl
	}

	return result
}
