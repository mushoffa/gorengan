package minio

import (
	"os"
	"strconv"
	"testing"
)

func getClient(t *testing.T) MinioService {
	URL := os.Getenv("URL")
	ACCESS_KEY := os.Getenv("ACCESS_KEY")
	SECRET_KEY := os.Getenv("SECRET_KEY")
	SSL, err := strconv.ParseBool(os.Getenv("SSL"))
	if err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(URL, ACCESS_KEY, SECRET_KEY, SSL)
	if err != nil {
		t.Fatal(err)
	}

	return client
}