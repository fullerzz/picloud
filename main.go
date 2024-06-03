package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/Kagami/go-avif"
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

// global var initialized before API to store info on the server's uploaded files
var uploadedFiles UploadedFiles

func loadConfig(confFileName string) {
	file, _ := os.Open(confFileName)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
}

func loadFileMetadata(metadataPath string) UploadedFiles {
	var files UploadedFiles
	if _, err := os.Stat(metadataPath); err == nil {
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &files)
		if err != nil {
			panic(err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// create the file if it doesn't exist
		_, err := os.Create(metadataPath)
		if err != nil {
			panic(err)
		}
	} else {
		// panic if error isn't caused by missing file
		panic(err)
	}

	return files
}

func writeFileMetadata() {
	data, err := json.Marshal(uploadedFiles)
	if err != nil {
		slog.Error("Error marshalling metadata")
		panic(err)
	}
	err = os.WriteFile("metadata.json", data, 0644)
	if err != nil {
		slog.Error("Error writing metadata")
		panic(err)
	}
}

func buildLink(rawFilename string) string {
	return fmt.Sprintf("http://pi.local:1234/file/%s", url.QueryEscape(rawFilename))
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

	metadata := &FileMetadata{Name: file.Filename, Tags: form.Value["tags"], Link: buildLink(file.Filename)}

	err = uploadFileToS3(metadata, buf.Bytes())
	if err != nil {
		slog.Error("Error uploading file to S3: %s", err)
		return err
	}

	// TODO: remove the following 3 lines once the writeMetadataToTable function is implemented fully
	uploadedFiles.Files = append(uploadedFiles.Files, *metadata)
	slog.Info("Updating file metadata")
	go writeFileMetadata()

	err = writeMetadataToTable(file.Filename)
	if err != nil {
		slog.Error("Error writing metadata to table: %v\n", err)
	}

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
	if file == nil {
		return c.String(http.StatusNotFound, "File not found")
	}
	file.Tags = append(file.Tags, tags.Tags...)
	return c.String(http.StatusOK, "File updated")
}

// e.GET("/file/:name", getFile)
func getFile(c echo.Context) error {
	encodedName := c.Param("name")
	name, err := url.QueryUnescape(encodedName)
	if err != nil {
		return err
	}

	avifFmt := c.QueryParam("avif")
	if avifFmt == "true" {
		return getAvif(c)
	} else {
		return c.File(fmt.Sprintf("%s%s", conf.FilePrefix, name))
	}
}

func getAvif(c echo.Context) error {
	encodedName := c.Param("name")
	name, err := url.QueryUnescape(encodedName)
	if err != nil {
		return err
	}

	// check if avif file already exists and return it if it does
	if _, err := os.Stat(fmt.Sprintf("%s%s.avif", conf.FilePrefix, name)); err == nil {
		slog.Info("Found existing AVIF file")
		return c.File(fmt.Sprintf("%s%s.avif", conf.FilePrefix, name))
	}

	// check if the src file exists
	if _, err := os.Stat(fmt.Sprintf("%s%s", conf.FilePrefix, name)); err != nil {
		slog.Info("File not found")
		return c.String(http.StatusNotFound, "File not found")
	}
	slog.Info("Src file found")

	// open the srcFile
	srcFile, err := os.Open(fmt.Sprintf("%s%s", conf.FilePrefix, name))
	if err != nil {
		return err
	}
	slog.Info("File opened")
	defer srcFile.Close()

	// create new avif file
	dstFile, err := os.Create(fmt.Sprintf("%s%s.avif", conf.FilePrefix, name))
	if err != nil {
		return err
	}
	slog.Debug(fmt.Sprintf("Created file %s%s.avif", conf.FilePrefix, name))

	// decode the src file
	img, err := jpeg.Decode(srcFile)
	if err != nil {
		slog.Error("Error decoding image")
		return err
	}

	// encode the img as avif file
	err = avif.Encode(dstFile, img, nil)
	if err != nil {
		slog.Error("Error encoding AVIF image")
		return err
	}
	slog.Debug("AVIF image encoded successfully")

	// return the file
	return c.File(fmt.Sprintf("%s%s.avif", conf.FilePrefix, name))
}

// e.GET("/files", listFiles)
func listFiles(c echo.Context) error {
	return c.JSON(http.StatusOK, uploadedFiles)
}

// e.GET("/files/search", searchFiles)
func searchFiles(c echo.Context) error {
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
	loadConfig("conf.json")
	// Load information about uploaded files
	uploadedFiles = loadFileMetadata("metadata.json")
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
	e.PATCH("/file/:name", updateFileTags)
	e.POST("/file/upload", saveFile)
	e.GET("/files", listFiles)
	e.GET("/files/search", searchFiles)

	e.Logger.Fatal(e.Start(":1234"))
}
