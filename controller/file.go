package controller

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jazaret/go-fileweb/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

type file struct {
	uploadTemplate *template.Template
	listTemplate   *template.Template
	uploadComplete *template.Template
}

var awsSession *session.Session
var (
	region     = "us-east-1"        //os.Getenv("AWS_REGION")
	bucketName = "jazar-testbucket" //os.Getenv("BUCKET_NAME")
)

func (f file) init() {
	log.Println("Calling init")
	var err error
	awsSession, err = session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}
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
		result := uploadFileToS3(buffer, header.Filename)

		data := model.FileResponse{ID: result}
		w.Header().Add("Content-Type", "text/html")
		f.uploadComplete.Execute(w, data)
	} else {
		w.Header().Add("Content-Type", "text/html")
		f.uploadTemplate.Execute(w, nil)
	}
}

func (f file) listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := getFiles(w, r)
	if err != nil {
		return
	}
	fileList := model.FileList{
		Files: files,
	}
	w.Header().Add("Content-Type", "text/html")
	f.listTemplate.Execute(w, fileList)
}

func (f file) listFilesAPI(w http.ResponseWriter, r *http.Request) {
	fileList, err := getFiles(w, r)
	if err != nil {
		return
	}

	jData, err2 := json.Marshal(fileList)
	if err2 != nil {
		log.Printf("Error on Marshal %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}
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

	result, err := s3.New(awsSession).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Printf("Error on GetObject %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}

	fileName := *result.Metadata["File-Name"]

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", *result.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))

	io.Copy(w, result.Body)

	defer result.Body.Close()
}

func getFiles(w http.ResponseWriter, r *http.Request) ([]model.File, error) {
	var fileList []model.File
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}

	resp, err := s3.New(awsSession).ListObjects(params)
	if err != nil {
		log.Printf("Failed to GetList: %s\n", err.Error())
		w.Write([]byte(err.Error()))
		return nil, err
	}

	for _, key := range resp.Contents {
		file := model.File{}
		file.ID = *key.Key
		file.Size = *key.Size
		file.Name = GetFileNameFromS3(*key.Key)
		fileList = append(fileList, file)
	}

	return fileList, nil
}

func downloadFileToClient(w http.ResponseWriter, r *http.Request) {
	keySet := strings.Split(r.URL.Path, "download/")

	if len(keySet) < 1 {
		log.Printf("Error - key not specified\n")
		w.Write([]byte("Error - key not specified"))
		return
	}

	key := keySet[1]

	result, err := s3.New(awsSession).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Printf("Error on GetObject %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}

	fileName := *result.Metadata["File-Name"]

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", *result.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))

	io.Copy(w, result.Body)

	defer result.Body.Close()
}

//putTagsOnS3Object - adds set tags to file
func putTagsOnS3Object(key, uuid, fileName string) {

	tags := []*s3.Tag{
		&s3.Tag{
			Key:   aws.String("uuid"),
			Value: aws.String(uuid),
		},
		&s3.Tag{
			Key:   aws.String("file-name"),
			Value: aws.String(fileName),
		},
	}

	_, err := s3.New(awsSession).PutObjectTagging(&s3.PutObjectTaggingInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Tagging: &s3.Tagging{
			TagSet: tags,
		},
	})

	if err != nil {
		log.Printf("Error on PutObjectTagging %s\n", err)
		log.Fatal(err)
	}
}

func uploadFileToS3(file []byte, fileName string) string {
	u1 := uuid.Must(uuid.NewV4()).String()
	size := int64(len(file))
	key := u1

	_, err := s3.New(awsSession).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(file),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(file)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		Metadata: map[string]*string{
			"file-name": aws.String(fileName),
			"uuid":      aws.String(u1),
		},
	})

	if err != nil {
		log.Printf("Error on PutObject %s\n", err)
		log.Fatal(err)
	}

	putTagsOnS3Object(key, u1, fileName)

	return u1
}

//GetFileNameFromS3 returns the name of the file from the tag value
func GetFileNameFromS3(key string) string {
	const FileNameTag = "file-name"
	result := ""
	params := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	tags, _ := s3.New(awsSession).GetObjectTagging(params)

	for _, v := range tags.TagSet {
		if *v.Key == FileNameTag {
			result = *v.Value
			return result
		}
	}

	return result
}
