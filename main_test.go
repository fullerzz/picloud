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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockedS3FilesBucket struct {
	mock.Mock
}

func (m *MockedS3FilesBucket) UploadObjectToS3(file *FileUpload, api S3PutObjectAPI) (string, error) {
	args := m.Called(file, api)
	return args.String(0), args.Error(1)
}

func (m *MockedS3FilesBucket) GetObjectFromS3(key string, api S3GetObjectAPI) ([]byte, error) {
	args := m.Called(key, api)
	return args.Get(0).([]byte), args.Error(1)
}

func TestLoadConfig(t *testing.T) {
	loadConfig("testdata/conf.json")
	if assert.NotEmpty(t, conf) {
		assert.Equal(t, "./uploads/", conf.FilePrefix)
	}
}

// func TestListFiles(t *testing.T) {
// 	e := echo.New()
// 	req := httptest.NewRequest(http.MethodGet, "/files", nil)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)
// 	assert.NoError(t, listFiles(c))
// 	require.Equal(t, http.StatusOK, rec.Code)
// }

func TestSaveFile(t *testing.T) {
	// test setup - mock S3 and setup test database
	loadConfig("testdata/conf.json")
	err := connectDatabase("file_metadata_test")
	if err != nil {
		fmt.Println("Error connecting to database")
		panic(err)
	}
	filesBucket := new(MockedS3FilesBucket)
	filesBucket.On("UploadObjectToS3", mock.Anything, mock.Anything).Return("baxter.jpg", nil)

	e := echo.New()

	// create a multipart form for test request
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	fw, _ := w.CreateFormField("name")
	_, err = fw.Write([]byte("baxter.jpg"))
	if err != nil {
		fmt.Println("Error writing form field")
		panic(err)
	}
	fw, _ = w.CreateFormField("size")
	_, err = fw.Write([]byte("12345"))
	if err != nil {
		fmt.Println("Error writing form field 'size'")
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
		fmt.Println("Error writing form field 'tags'")
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

// func TestUpdateFileTagsNotFound(t *testing.T) {
// 	e := echo.New()

// 	tags := Tags{Tags: []string{"dog", "baxter"}}
// 	body, err := json.Marshal(tags)
// 	if err != nil {
// 		fmt.Println("Error marshalling tags")
// 		panic(err)
// 	}

// 	req := httptest.NewRequest(http.MethodPatch, "/file/baxter.jpg", bytes.NewBuffer(body))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)

// 	assert.NoError(t, updateFileTags(c))
// 	assert.Equal(t, "File not found", rec.Body.String())
// 	require.Equal(t, http.StatusNotFound, rec.Code)
// }
