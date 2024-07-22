package minio

import (
	"bytes"
	"encoding/csv"
	"image"
	"os"
	"testing"

	img "github.com/mushoffa/gorengan/image"
	"github.com/stretchr/testify/assert"
)

func TestGetObject_Image_Success(t *testing.T) {
	client := getClient(t)

	object, err := client.GetObject("test", "input_image.png", "image")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, object)
	assert.Nil(t, err)

	data, _, _ := image.Decode(bytes.NewReader(object))
	writeErr := img.WriteFile(data, "./test", "get_input_image")
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}

func TestGetObject_CSV_Success(t *testing.T) {
	client := getClient(t)

	object, err := client.GetObject("test", "input_csv.csv", "csv")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, object)
	assert.Nil(t, err)

	records, readErr := csv.NewReader(bytes.NewBuffer(object)).ReadAll()
	if readErr != nil {
		t.Fatal(readErr)
	}

	file, createErr := os.Create("test/get_input_csv.csv")
	if createErr != nil {
		t.Fatal(createErr)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	writeErr := writer.WriteAll(records)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
}