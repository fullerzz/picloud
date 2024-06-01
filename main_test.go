package main

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	loadConfig("testdata/conf.json")
	if assert.NotEmpty(t, conf) {
		assert.Equal(t, "uploads/", conf.FilePrefix)
	}
}

func TestLoadFileMetadata(t *testing.T) {
	expectedFileMetadata := FileMetadata{Name: "baxter.jpg", Tags: []string{}, Link: "baxter.jpg"}
	expected := UploadedFiles{Files: []FileMetadata{expectedFileMetadata}}
	uploadedFiles = loadFileMetadata("testdata/metadata.json")
	if assert.NotEmpty(t, uploadedFiles) {
		assert.Equal(t, expected, uploadedFiles)
	}
}

func TestLoadMissingFileMetadata(t *testing.T) {
	testMetadataPath := "testdata/nonexistant_metadata.json"
	expected := UploadedFiles{}
	uploadedFiles = loadFileMetadata(testMetadataPath)
	if assert.NotNil(t, uploadedFiles) {
		assert.Equal(t, expected, uploadedFiles)
	}
	// cleanup
	err := os.Remove(testMetadataPath)
	if err != nil {
		fmt.Println("Error removing test metadata file")
		panic(err)
	}
}

func TestListFiles(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.NoError(t, listFiles(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestSaveFile(t *testing.T) {
	e := echo.New()

	// create a multipart form for test request
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	fw, _ := w.CreateFormField("name")
	_, err := fw.Write([]byte("baxter.jpg"))
	if err != nil {
		fmt.Println("Error writing form field")
		panic(err)
	}
	fw, _ = w.CreateFormField("size")
	_, err = fw.Write([]byte("12345"))
	if err != nil {
		fmt.Println("Error writing form field")
		panic(err)
	}
	fw, _ = w.CreateFormFile("file", "baxter.jpg")
	// load baxter.jpg into memory
	testFilePath := "testdata/baxter.jpg"
	fileData, err := os.ReadFile(testFilePath)
	if err != nil {
		fmt.Printf("Error reading file %s\n", testFilePath)
		panic(err)
	}
	_, err = fw.Write(fileData)
	if err != nil {
		fmt.Println("Error writing image data to form file")
		panic(err)
	}
	fw, _ = w.CreateFormField("tags")
	_, err = fw.Write([]byte("dog, baxter"))
	if err != nil {
		fmt.Println("Error writing form field")
		panic(err)
	}
	w.Close()

	// setup test request and recorder
	req := httptest.NewRequest(http.MethodPost, "/file/upload", buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	assert.NoError(t, saveFile(c))
	require.Equal(t, http.StatusOK, rec.Code)
}
