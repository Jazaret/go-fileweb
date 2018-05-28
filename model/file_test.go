package model

import "testing"

func TestGetFileEmptyKeyShouldFail(t *testing.T) {
	testRepo := new(mockRepo)
	fileRepo = testRepo

	_, err := GetFileFromRepo("")

	if err == nil || err.Error() != "Empty or invalid key" {
		t.Errorf("Empty string should not be allowed")
	}
}

func TestGetFileInvalidKeyShouldFail(t *testing.T) {
	testRepo := new(mockRepo)
	fileRepo = testRepo

	_, err := GetFileFromRepo("2344")

	if err == nil || err.Error() != "Empty or invalid key" {
		t.Errorf("Invalid key should not be allowed")
	}
}

func TestGetFileValidKeyShouldPass(t *testing.T) {
	testRepo := new(mockRepo)
	fileRepo = testRepo

	_, err := GetFileFromRepo("6d468b76-cf62-4b90-a238-bc0c4ace1648")

	if err != nil {
		t.Errorf("Valid key should pass")
	}
}

func TestUploadFileWithNoNameShouldFail(t *testing.T) {
	testRepo := new(mockRepo)
	fileRepo = testRepo
	var file []byte

	_, _, err := UploadFileToRepo(file, "")

	if err == nil || err.Error() != "File name is required" {
		t.Errorf("Invalid file should not be allowed")
	}
}

func TestUploadFileWithNoNameShouldPass(t *testing.T) {
	testRepo := new(mockRepo)
	fileRepo = testRepo
	var file []byte

	_, _, err := UploadFileToRepo(file, "test")

	if err != nil {
		t.Errorf("Valid file name should pass")
	}
}

func TestGetListShouldPass(t *testing.T) {
	testRepo := new(mockRepo)
	fileRepo = testRepo

	_, err := GetFiles()

	if err != nil {
		t.Errorf("GetFiles failed with error %s", err)
	}
}
