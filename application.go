package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

var index = `<html>
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

func ReceiveFile(w http.ResponseWriter, r *http.Request) string {
	// in your case file would be fileupload
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error on FormFile %s\n", err)
		log.Fatal(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")

	fmt.Printf("File name %s\n", name[0])
	buffer, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Error on ReadAll %s\n", err)
		log.Fatal(err)
	}
	result := UploadFileToS3(buffer, header.Filename)

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

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			log.Printf("HTTP POST received on upload...\n")
			id := ReceiveFile(w, r)
			w.Header().Set("Content-Type", "application/json")
			data := fileResponse{ID: id}
			jData, err := json.Marshal(data)
			if err != nil {
				log.Printf("Error on Marshal %s\n", err)
				log.Fatal(err)
				return
			}
			w.Write(jData)
		} else {
			log.Printf("Serving %s to %s...\n", indexPage, r.RemoteAddr)
			w.Write([]byte(index))
		}
	})
	log.Printf("Listening on port %s\n\n", port)
	http.ListenAndServe(":"+port, nil)
}
