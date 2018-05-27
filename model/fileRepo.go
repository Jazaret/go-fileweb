package model

import (
	"bytes"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var fileRepo FileRepository

//FileRepository - interface to handle file repo methods
type FileRepository interface {
	GetObject(key string) (File, error)
	PutObject(key string, file []byte, fileName string) error
	ListObjects() ([]File, error)
}

type s3Repository struct {
	s3repo     *s3.S3
	bucketName string
}

func (s s3Repository) GetObject(key string) (File, error) {
	file := &File{}
	result, err := s.s3repo.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return *file, err
	}

	file.Name = *result.Metadata["File-Name"]
	file.ContentType = *result.ContentType
	file.Size = *result.ContentLength
	file.Blob = result.Body

	return *file, nil
}

func (s s3Repository) PutObject(key string, file []byte, fileName string) error {
	size := int64(len(file))
	_, err := s.s3repo.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(s.bucketName),
		Key:                  aws.String(key),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(file),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(file)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		Metadata: map[string]*string{
			"file-name": aws.String(fileName),
			"uuid":      aws.String(key),
		},
	})

	s.putTagsOnS3Object(key, key, fileName)

	return err
}

func (s s3Repository) ListObjects() ([]File, error) {

	var fileList []File

	params := &s3.ListObjectsInput{
		Bucket: aws.String(s.bucketName),
	}
	resp, _ := s.s3repo.ListObjects(params)

	for _, key := range resp.Contents {
		file := File{}
		file.ID = *key.Key
		file.Size = *key.Size
		file.Name = s.GetFileNameFromS3(*key.Key)
		fileList = append(fileList, file)
	}

	return fileList, nil
}

//InitRepository - initializes the file repository with a repo connection
func InitRepository(region string, bucketName string) {
	log.Println("Calling init")
	var err error
	awsSession, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		log.Fatal(err)
	}
	fileRepo = &s3Repository{s3.New(awsSession), bucketName}
}

//putTagsOnS3Object - adds set tags to file
func (s s3Repository) putTagsOnS3Object(key, uuid, fileName string) {

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

	_, err := s.s3repo.PutObjectTagging(&s3.PutObjectTaggingInput{
		Bucket: aws.String(s.bucketName),
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

//GetFileNameFromS3 returns the name of the file from the tag value
func (s s3Repository) GetFileNameFromS3(key string) string {
	const FileNameTag = "file-name"
	result := ""
	params := &s3.GetObjectTaggingInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	tags, _ := s.s3repo.GetObjectTagging(params)

	for _, v := range tags.TagSet {
		if *v.Key == FileNameTag {
			result = *v.Value
			return result
		}
	}

	return result
}
