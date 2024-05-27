package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	loadConfig()
	if assert.NotEmpty(t, conf) {
		assert.Equal(t, "/opt/picloud/uploads/", conf.FilePrefix)
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
