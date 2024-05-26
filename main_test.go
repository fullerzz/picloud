package main

import (
	"net/http"
	"net/http/httptest"
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
	// TODO: Delete testdata/nonexistant_metadata.json if it exists before this test is ran
	expected := UploadedFiles{Files: []FileMetadata{}}
	uploadedFiles = loadFileMetadata("testdata/nonexistant_metadata.json")
	if assert.NotNil(t, uploadedFiles) {
		// TODO:  Fix following assertion. Actual type being returned is an interface I think..
		assert.Equal(t, expected, uploadedFiles)
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
