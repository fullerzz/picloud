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

type Tags struct {
	Tags []string `json:"tags"`
}

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

// global var initialized before API to store info on the server's uploaded files
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

// buildLink returns a link to be used in the FileMetadata struct on initialization
func buildLink(rawFilename string) string {
	return fmt.Sprintf("http://localhost:1234/file/%s", url.QueryEscape(rawFilename))
}

// e.POST("/file/upload", saveFile)
func saveFile(c echo.Context) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	// Extract file from form
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
	// Extract tags from form
	tags := form.Value["tags"]
	uploadedFiles.Files = append(uploadedFiles.Files, FileMetadata{Name: file.Filename, Tags: tags, Link: buildLink(file.Filename)})
	go writeFileMetadata()
	return c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully!", file.Filename))
}

// e.PATCH("/file/:name", updateFileTags)
func updateFileTags(c echo.Context) error {
	// extract tags from request
	var tags Tags
	err := c.Bind(&tags)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid request")
	}
	// decode the name
	encodedName := c.Param("name")
	name, err := url.QueryUnescape(encodedName)
	if err != nil {
		return err
	}
	// get the file from the uploadedFiles
	var file *FileMetadata
	for i, f := range uploadedFiles.Files {
		if f.Name == name {
			file = &uploadedFiles.Files[i]
			break
		}
	}
	file.Tags = append(file.Tags, tags.Tags...)
	return c.String(http.StatusOK, "File updated")
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
	return c.File(name)
}

// e.GET("/files", listFiles)
func listFiles(c echo.Context) error {
	// list all available files
	return c.JSON(http.StatusOK, uploadedFiles)
}

// e.GET("/files/search", searchFiles)
func searchFiles(c echo.Context) error {
	// search for files by tag
	// get the tag from the request
	tag := c.QueryParam("tag")
	// search for the tag in the uploadedFiles
	var foundFiles []FileMetadata
	for _, file := range uploadedFiles.Files {
		for _, t := range file.Tags {
			if t == tag {
				foundFiles = append(foundFiles, file)
				break
			}
		}
	}
	if len(foundFiles) == 0 {
		return c.NoContent(http.StatusNoContent)
	} else {
		return c.JSON(http.StatusOK, foundFiles)
	}
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
	e.PATCH("/file/:name", updateFileTags)
	e.POST("/file/upload", saveFile)
	e.GET("/files", listFiles)
	e.GET("/files/search", searchFiles)

	e.Logger.Fatal(e.Start(":1234"))
}
