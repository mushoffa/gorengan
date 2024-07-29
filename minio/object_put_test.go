package minio

import (
	"os"
	"testing"

	"github.com/mushoffa/gorengan/image"
	"github.com/stretchr/testify/assert"
)

func TestPutObject_Image_Success(t *testing.T) {
	input, readErr := image.ReadFile("./test/put_input_image.png")
	if readErr != nil {
		t.Fatal(readErr)
	}

	data, byteErr := image.New(input).Bytes()
	if byteErr != nil {
		t.Fatal(byteErr)
	}

	client := getClient(t)
	putErr := client.PutObject("test", "put_input_image.png", data, "image")

	assert.Nil(t, putErr)
}

func TestPutObject_CSV_Success(t *testing.T) {
	data, errRead := os.ReadFile("test/put_input_csv.csv")
	if errRead != nil {
		t.Fatal(errRead)
	}

	client := getClient(t)
	putErr := client.PutObject("test", "put_input_csv.csv", data, "csv")

	assert.Nil(t, putErr)
}
