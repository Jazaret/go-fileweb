package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

var awsSession *session.Session

type fileResponse struct {
	ID string `json: ID`
}

//File is the main data struct of our file system
type File struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
	Size int64  `json:"Size"`
}

const index = `<html>
<head>
   	<title>Upload file</title>
</head>
<body>
<form enctype="multipart/form-data" action="/upload" method="post">
	<input type="file" name="file" />
	<input type="hidden" name="token" value="{{.}}"/>
	<input type="submit" value="upload" />
</form>
</body>
</html>`

var (
	region     = "us-east-1"        //os.Getenv("AWS_REGION")
	bucketName = "jazar-testbucket" //os.Getenv("BUCKET_NAME")
)

//PutTagsOnS3Object - adds set tags to file
func PutTagsOnS3Object(key, uuid, fileName string) {

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

func UploadFileToS3(file []byte, fileName string) string {
	key := fileName
	u1 := uuid.Must(uuid.NewV4()).String()
	size := int64(len(file))

	//var buff bytes.Buffer
	//fileSize, err := buff.ReadFrom(file)

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

	PutTagsOnS3Object(key, u1, fileName)

	return u1
}

//GetFileFromS3 gets the file from s3 using GetObject
func GetFileFromS3(fileDir string) {
	key := fileDir

	result, err := s3.New(awsSession).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out, err := os.Create(fileDir)

	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	io.Copy(out, result.Body)
	result.Body.Close()
}

func ReceiveFileFromClient(w http.ResponseWriter, r *http.Request) {

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
	result := UploadFileToS3(buffer, header.Filename)

	w.Header().Set("Content-Type", "application/json")
	data := fileResponse{ID: result}
	jData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error on Marshal %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(jData)
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path
	if true {
		key = "upload2.txt"
	}

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
	//uuid := *result.Metadata["Uuid"]

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", *result.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(*result.ContentLength, 10))

	io.Copy(w, result.Body)

	defer result.Body.Close()
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	var fileList []File
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}

	resp, err := s3.New(awsSession).ListObjects(params)
	if err != nil {
		log.Printf("Failed to GetList: %s\n", err.Error())
		w.Write([]byte(err.Error()))
		return
	}

	for _, key := range resp.Contents {
		file := File{}
		file.ID = *key.Key
		file.Size = *key.Size
		file.Name = GetFileNameFromS3(*key.Key)
		fileList = append(fileList, file)
		fmt.Println(*key.Key)
	}

	jData, err := json.Marshal(fileList)
	if err != nil {
		log.Printf("Error on Marshal %s\n", err)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(jData)
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

func init() {
	log.Println("Calling init")
	var err error
	awsSession, err = session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	f, _ := os.Create("/var/log/golang/golang-server.log")
	defer f.Close()
	log.SetOutput(f)

	const indexPage = "public/index.html"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
		w.Write([]byte(index))
	})

	http.HandleFunc("/list", listFiles)

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			log.Printf("HTTP POST received on upload...\n")
			ReceiveFileFromClient(w, r)
		} else {
			log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
			w.Write([]byte(index))
		}
	})

	http.HandleFunc("/download/", downloadFile)

	log.Printf("Listening on port %s\n\n", port)
	http.ListenAndServe(":"+port, nil)
}
