package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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

type Configuration struct {
	FilePrefix string
}

var conf Configuration
var metadataTable MetadataTable
var filesBucket S3FilesBucket

func loadConfig(confFileName string) {
	file, _ := os.Open(confFileName)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
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

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, src); err != nil {
		slog.Error("Error copying file: %s", err)
		return err
	}

	fileUpload := &FileUpload{Name: file.Filename, Size: len(buf.Bytes()), Content: buf.Bytes(), Tags: form.Value["tags"]}

	err = uploadFileToS3(fileUpload)
	if err != nil {
		slog.Error("Error uploading file to S3: %s", err)
		return err
	}

	err = writeMetadataToTable(fileUpload)
	if err != nil {
		slog.Error("Error writing metadata to table", "err", err)
		return err
	}

	return c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully!", file.Filename))
}

// e.GET("/file/:name", getFile)
func getFile(c echo.Context) error {
	encodedName := c.Param("name")
	name, err := url.QueryUnescape(encodedName)
	if err != nil {
		return err
	}
	files, err := metadataTable.Query(name)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, files)
}

// e.GET("/files", listFiles)
// func listFiles(c echo.Context) error {
// 	return c.JSON(http.StatusOK, uploadedFiles)
// }

// e.GET("/files/search", searchFiles)
// func searchFiles(c echo.Context) error {
// 	tag := c.QueryParam("tag")
// 	// search for the tag in the uploadedFiles
// 	var foundFiles []FileMetadata
// 	for _, file := range uploadedFiles.Files {
// 		for _, t := range file.Tags {
// 			if t == tag {
// 				foundFiles = append(foundFiles, file)
// 				break
// 			}
// 		}
// 	}
// 	if len(foundFiles) == 0 {
// 		return c.NoContent(http.StatusNoContent)
// 	} else {
// 		return c.JSON(http.StatusOK, foundFiles)
// 	}
// }

func main() {
	loadConfig("conf.json")
	createClientConnections()
	e := echo.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.CORS())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/file/:name", getFile)
	// e.PATCH("/file/:name", updateFileTags)
	e.POST("/file/upload", saveFile)
	// e.GET("/files", listFiles)
	// e.GET("/files/search", searchFiles)

	e.Logger.Fatal(e.Start(":1234"))
}
