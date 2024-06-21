package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	b64 "encoding/base64"
	"fmt"
	"io"
	"time"
)

type FileMetadataRecord struct {
	FileName        string  `json:"file_name"`
	ObjectKey       string  `json:"object_key"`
	Sha256          string  `json:"file_sha256"`
	UploadTimestamp int64   `json:"upload_timestamp"`
	Tags            string  `json:"tags"` // comma separated tags
	LocalPath       *string `json:"local_path"`
	CacheTimestamp  int64   `json:"cache_timestamp"`
}

var DB *sql.DB

func connectDatabase() error {
	db, err := sql.Open("sqlite3", "./metadata.db")
	if err != nil {
		return err
	}

	DB = db
	return nil
}

func addMetadataToTable(metadata *FileMetadataRecord) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO file_metadata (file_name, object_key, file_sha256, upload_timestamp, tags, local_path, cache_timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(metadata.FileName, metadata.ObjectKey, metadata.Sha256, metadata.UploadTimestamp, metadata.Tags, metadata.LocalPath, metadata.CacheTimestamp)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func updateCacheDetailsInTable(metadata *FileMetadataRecord) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("UPDATE file_metadata SET local_path = ?, cache_timestamp = ? WHERE file_name = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(*metadata.LocalPath, metadata.CacheTimestamp, metadata.FileName)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func listFilesInTable() ([]FileMetadataRecord, error) {
	rows, err := DB.Query("SELECT file_name, object_key, file_sha256, upload_timestamp, tags, local_path, cache_timestamp FROM file_metadata")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileMetadataRecord
	for rows.Next() {
		var file FileMetadataRecord
		err = rows.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags, &file.LocalPath, &file.CacheTimestamp)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func getFileMetadataFromTable(filename string) (*FileMetadataRecord, error) {
	row := DB.QueryRow("SELECT file_name, object_key, file_sha256, upload_timestamp, tags FROM file_metadata WHERE file_name = ?", filename)
	var file FileMetadataRecord
	err := row.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func queryTags(tagName string) ([]FileMetadataRecord, error) {
	rows, err := DB.Query("SELECT file_name, object_key, file_sha256, upload_timestamp, tags FROM file_metadata WHERE tags LIKE ?", "%"+tagName+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileMetadataRecord
	for rows.Next() {
		var file FileMetadataRecord
		err = rows.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func getSha256Checksum(fileContent *[]byte) string {
	h := sha256.New()
	_, err := io.Copy(h, bytes.NewReader(*fileContent))
	if err != nil {
		fmt.Println("Error calculating copying bytes in getSha256Checksum")
		panic(err)
	}
	checksum := b64.StdEncoding.EncodeToString(h.Sum(nil))
	return checksum
}

func getTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}
