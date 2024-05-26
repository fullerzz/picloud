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
	// TODO: inject mock data to load for this test and assert mocked data is returned
	expectedFileMetadata := FileMetadata{Name: "baxter.jpg", Tags: []string{}, Link: "baxter.jpg"}
	expected := UploadedFiles{Files: []FileMetadata{expectedFileMetadata}}
	uploadedFiles = loadFileMetadata("testdata/metadata.json")
	if assert.NotEmpty(t, uploadedFiles) {
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
