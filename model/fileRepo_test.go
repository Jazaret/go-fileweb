package model

type mockRepo struct {
}

func (s *mockRepo) GetObject(key string) (File, error) {
	return File{}, nil
}

func (s *mockRepo) PutObject(key string, file []byte, fileName string) (string, error) {
	return "", nil
}

func (s *mockRepo) ListObjects() ([]File, error) {
	var files []File
	return files, nil
}
