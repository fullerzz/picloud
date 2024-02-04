package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type FileUpload struct {
	Name    string   `json:"name"`
	Size    int      `json:"size"`
	Content []byte   `json:"content"`
	Tags    []string `json:"tags"`
}

type FileMetadata struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
	Link string   `json:"link"`
}

type UploadedFiles struct {
	Files []FileMetadata `json:"files"`
}

var uploadedFiles UploadedFiles

func loadFileMetadata() UploadedFiles {
	// load file metadata from file
	data, err := os.ReadFile("metadata.json")
	if err != nil {
		panic(err)
	}
	var files UploadedFiles
	json.Unmarshal(data, &files)
	return files
}

func writeFileMetadata() {
	// write file metadata to file
	data, err := json.Marshal(uploadedFiles)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("metadata.json", data, 0644)
	if err != nil {
		panic(err)
	}
}

// e.POST("/file/upload", saveFile)
func saveFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Destination
	dst, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// TODO: Update Tags and Link fields
	uploadedFiles.Files = append(uploadedFiles.Files, FileMetadata{Name: file.Filename, Tags: []string{}, Link: file.Filename})
	go writeFileMetadata()

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully with fields: %s!!</p>", file.Filename, c.FormValue("title")))
}

// e.GET("/file/:name", getFile)
func getFile(c echo.Context) error {
	// name string will be urlencoded
	encodedName := c.Param("name")
	// decode the name
	name, err := url.QueryUnescape(encodedName)
	if err != nil {
		return err
	}
	// load file into memory and return
	file, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	return c.File(string(file))
}

func listFiles(c echo.Context) error {
	// list all available files
	return c.JSON(http.StatusOK, uploadedFiles)
}

func main() {
	// Load information about uploaded files
	uploadedFiles = loadFileMetadata()
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/file/:name", getFile)
	e.POST("/file/upload", saveFile)
	e.GET("/files", listFiles)

	e.Logger.Fatal(e.Start(":1234"))
}
