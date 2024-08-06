package minio

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioService lists currently available API
type MinioService interface {
	GetInstance() *minio.Client
	CreateBucket(string, string) error
	ListBuckets() ([]minio.BucketInfo, error)
	DeleteBucket(string) error
	ListFolders(string, ...string) (string, error)
	PutObject(string, string, []byte, ...string) error
	GetObject(string, string, ...string) ([]byte, error)
}

// client implements [MinioService] interface.
type client struct {
	instance *minio.Client
}

// NewClient returns a new MinIO client which exposes [MinioService] methods.
func NewClient(url, accessKey, secretKey string, useSSL bool) (MinioService, error) {
	minioClient, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &client{minioClient}, nil
}

// GetInstance returns [minio.Client] object.
func (c *client) GetInstance() *minio.Client {
	return c.instance
}

// CreateBucket creates a new bucket with given name.
func (c *client) CreateBucket(bucketName, region string) error {
	return c.instance.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
}

// ListBuckets returns a list of buckets created previously.
func (c *client) ListBuckets() ([]minio.BucketInfo, error) {
	return c.instance.ListBuckets(context.Background())
}

// DeleteBucket deletes a bucket created previously.
func (c *client) DeleteBucket(bucketName string) error {
	return c.instance.RemoveBucket(context.Background(), bucketName)
}

// ListFolders returns a list of foler path and files in a bucket.
func (c *client) ListFolders(bucketName string, paths ...string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	folderPath := c.getPath(paths...)

	var payload []map[string]string

	objects := c.listObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: folderPath})

	for object := range objects {
		if object.Err != nil {
			return "", object.Err
		}

		name := strings.Split(object.Key, "/")

		obj := map[string]string{
			"id":        object.ETag,
			"name":      name[1],
			"timestamp": object.LastModified.String(),
		}

		payload = append(payload, obj)
	}

	j, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// PutObject uploads object to your MinIO server.
func (c *client) PutObject(bucketName, fileName string, object []byte, paths ...string) error {
	filePath := c.getPath(paths...) + fileName
	reader := bytes.NewReader(object)
	size := int64(len(object))

	_, err := c.instance.PutObject(context.Background(), bucketName, filePath, reader, size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}

	return nil
}

// GetObject returns object or file from your MinIO server.
func (c *client) GetObject(bucketName, fileName string, paths ...string) ([]byte, error) {
	objectPath := c.getPath(paths...) + fileName

	object, err := c.instance.GetObject(context.Background(), bucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	reader, readErr := io.ReadAll(object)
	if readErr != nil {
		return nil, err
	}

	return reader, nil
}

// getPath
func (c *client) getPath(paths ...string) string {
	var folderPath string

	length := len(paths)

	if length > 0 {

		for _, path := range paths {
			folderPath += path + "/"
		}
	}

	return folderPath
}

// listObjects
func (c *client) listObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	return c.instance.ListObjects(ctx, bucketName, opts)
}
