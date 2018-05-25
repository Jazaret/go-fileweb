package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

var (
	region     = os.Getenv("AWS_REGION")
	bucketName = os.Getenv("BUCKET_NAME")
)

var awsSession *session.Session

type fileResponse struct {
	ID string `json: ID`
}

//PutTagsOnS3Object - adds set tags to file
func PutTagsOnS3Object(key, uuid, fileDir string) {

	tags := []*s3.Tag{
		&s3.Tag{
			Key:   aws.String("uuid"),
			Value: aws.String(uuid),
		},
		&s3.Tag{
			Key:   aws.String("file-name"),
			Value: aws.String("doiwork"),
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
		log.Fatal(err)
	}
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func AddFileToS3(s *session.Session, fileDir string) (id string, err error) {
	key := fileDir
	u1 := uuid.Must(uuid.NewV4()).String()

	// Open the file for use
	file, err := os.Open(fileDir)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		Metadata: map[string]*string{
			"file-name": aws.String(fileDir),
			"uuid":      aws.String(u1),
		},
	})

	PutTagsOnS3Object(key, u1, fileDir)

	return u1, err
}

func ReceiveFile(w http.ResponseWriter, r *http.Request) {
	var Buf bytes.Buffer
	// in your case file would be fileupload
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")

	fmt.Printf("File name %s\n", name[0])
	// Copy the file data to my buffer
	io.Copy(&Buf, file)
	// do something with the contents...
	// I normally have a struct defined and unmarshal into a struct, but this will
	// work as an example
	contents := Buf.String()
	fmt.Println(contents)
	// I reset the buffer in case I want to use it again
	// reduces memory allocations in more intense projects
	Buf.Reset()
	// do something else
	// etc write header
	return
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
		http.ServeFile(w, r, indexPage)
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			ReceiveFile(w, r)
			w.Header().Set("Content-Type", "application/json")
			data := fileResponse{ID: "1123"}
			jData, err := json.Marshal(data)
			if err != nil {
				panic(err)
				return
			}
			w.Write(jData)
		} else {
			log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
			http.ServeFile(w, r, indexPage)
		}
	})
	log.Printf("Listening on port %s\n\n", port)
	http.ListenAndServe(":"+port, nil)
}
