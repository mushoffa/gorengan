# MinIO SDK
## Official Documentation
* [Quickstart Guide](https://min.io/docs/minio/linux/developers/go/minio-go.html)
* [Github](https://github.com/minio/minio-go)

## API
* GetInstance
* CreateBucket
* ListBuckets
* DeleteBucket
* ListFolders
* PutObject
* GetObject
```go
// MinioService lists currently available API 
type MinioService interface {
	GetInstance() *minio.Client
	CreateBucket(string, string) error
	ListBuckets() ([]minio.BucketInfo, error)
	DeleteBucket(string) error
	ListFolders(string, ...string) ([]string, error)
	PutObject(string, string, []byte, ...string) error
	GetObject(string, string, ...string) ([]byte, error)
}
```

## Usage
Use it right away with pre-defined API service.
1. Parse the ```MinioService``` interface to your function parameter.
```go
package repository

import (
	"context"

	"github.com/mushoffa/gorengan/minio"
)

var YOUR_MINIO_REGION = "region"

type Repository struct {
	storage minio.MinioService
}

func NewRepository(storage minio.MinioService) *Repository {
	return &Repository{storage:storage}
}

func (r *Repository) CreateBucket(bucketName string) error {
	return r.storage.CreateBucket(bucketName, YOUR_MINIO_REGION)
}

func (r *Repository) ListBuckets() ([]string, error) {
	// Or you can access minio client instance and perform your own process
	objects, err := r.storage.GetInstance().ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}

	var buckets []string

	for _, object := range objects {
		buckets = append(buckets, object.Name)
	}

	return buckets, nil
}
```
2. Instantiate the minio client and parse it to your defined function parameter.
```go
package main

import (
	"YOUR_PROJECT/repository"

	"github.com/mushoffa/gorengan/minio"
)

var (
	URL = "YOUR_MINIO_URL"
	ACCESS_KEY = "YOUR_MINIO_ACCESS_KEY"
	SECRET_KEY = "YOUR_MINIO_SECRET_KEY"
	SSL = true
)

func main() {
	client, err := minio.NewClient(URL, ACCESS_KEY, SECRET_KEY, SSL)
	if err != nil {
		panic(err)
	}

	r := repository.NewRepository(client)

	// Do your stuff here

	...
}
```

## Extension
Embed ```MinioService``` interface to your client struct or both your client struct and interface.

```go
package infrastructure

import (
	"github.com/mushoffa/gorengan/minio"
)

type YourMinioService interface {
	minio.MinioService // Optional, embed if want to access exposed API
	YourMinioGetFunction()
	YourMinioPutFunction()
	YourMinioSpecificFunction()
}

type YourMinioClient struct {
	minio.MinioService
}

func NewMinioClient(url, accessKey, secretKey string, useSSL bool) (YourMinioService, error) {
	client, err := minio.NewClient(url, accessKey, secretKey, useSSL)
	if err != nil {
		return nil, err
	}

	your_minio_client := &YourMinioClient{client}

	return your_minio_client, nil
}

func (c *YourMinioClient) YourMinioGetFunction() {
	object, err := c.GetObject(...)
	
	// Do your stuff after receiving object
	...
}

func (c *YourMinioClient) YourMinioPutFunction() {
	// Do your stuff before sending object
	...

	c.PutObject(...)
}

func (c *YourMinioClient) YourMinioSpecificFunction() {
	client := c.GetInstance()

	// Do your stuff here
	...
}
```
## Testing
### Prerequisite
Export your MinIO environment variables.
```shell
export URL=YOUR_MINIO_URL ACCESS_KEY=YOUR_MINIO_ACCESS_KEY SECRET_KEY=YOUR_MINIO_SECRET_KEY SSL=false
```
### Setup
Setup your MinIO API variable such as bucket name, file name, or file path on the test file(s):
* [bucket_test.go](./bucket_test.go)
* [object_get_test.go](./object_get_test.go)
* [object_put_test.go](./object_put_test.go)

### Run
Run test file(s) on terminal or using [Makefile](./Makefile).

#### Terminal
```shell
mushoffa/goreangan/minio$ go test -v
```
#### Makefile
* Run all tests
```shell
mushoffa/goreangan/minio$ make test-all
```
* Run [bucket_test.go](./bucket_test.go)
```shell
mushoffa/goreangan/minio$ make test-bucket
```
* Run [object_get_test.go](./object_get_test.go) and [object_put_test.go](./object_put_test.go)
```shell
mushoffa/goreangan/minio$ make test-object
```