package model

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

var fileRepo FileRepository

//FileRepository - interface to handle file repo methods
type FileRepository interface {
	GetObject(key string) (File, error)
	PutObject(key string, file []byte, fileName string) (string, error)
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

	accessTokenExpiryDate, timeErr := time.Parse(time.RFC3339, *result.Metadata["Access-Token-Expry-Date"])
	if timeErr != nil {
		return *file, timeErr
	}
	if time.Now().UTC().After(accessTokenExpiryDate) {
		return *file, errors.New("Access time expired")
	}

	file.Name = *result.Metadata["File-Name"]
	file.ContentType = *result.ContentType
	file.Size = *result.ContentLength
	file.Blob = result.Body

	return *file, nil
}

func (s s3Repository) PutObject(key string, file []byte, fileName string) (string, error) {
	size := int64(len(file))
	accessToken := uuid.Must(uuid.NewV4()).String()
	accessTokenExpiryDate := time.Now().UTC().AddDate(0, 0, 7)
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
			"file-name":               aws.String(fileName),
			"access-token":            aws.String(accessToken),
			"access-token-expry-date": aws.String(accessTokenExpiryDate.Format(time.RFC3339)),
		},
	})

	if err != nil {
		return "", err
	}

	accessToken, err2 := s.putTagsOnS3Object(key, fileName, accessToken, accessTokenExpiryDate)
	if err2 != nil {
		return "", err
	}

	return accessToken, nil
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

//putTagsOnS3Object - adds set tags to file. Adds
func (s s3Repository) putTagsOnS3Object(key, fileName, accessToken string, accessTokenExpiryDate time.Time) (string, error) {

	tags := []*s3.Tag{
		&s3.Tag{
			Key:   aws.String("file-name"),
			Value: aws.String(fileName),
		},
		&s3.Tag{
			Key:   aws.String("access-token"),
			Value: aws.String(accessToken),
		},
		&s3.Tag{
			Key:   aws.String("access-token-expry-date"),
			Value: aws.String(accessTokenExpiryDate.Format(time.RFC3339)),
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
		return "", err
	}

	return accessToken, nil
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
