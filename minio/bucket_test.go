package minio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket_Create_Success(t *testing.T) {
	client := getClient(t)
	err := client.CreateBucket("test", "local")
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, err)
}

func TestBucket_List_Success(t *testing.T) {
	client := getClient(t)
	buckets, err := client.ListBuckets()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, buckets)
	assert.Nil(t, err)

	t.Log(buckets)
}

func TestBucket_ListFolders_Success(t *testing.T) {
	client := getClient(t)
	folders, err := client.ListFolders("test")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, folders)
	assert.Nil(t, err)

	t.Log(folders)
}