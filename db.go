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
	Link            string  `json:"link"`
}

var DB *sql.DB

func connectDatabase(tableName string) error {
	db, err := sql.Open("sqlite3", fmt.Sprintf("./%s.db", tableName))
	if err != nil {
		return err
	}

	DB = db
	_, err = DB.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (file_name TEXT NOT NULL PRIMARY KEY UNIQUE, object_key TEXT UNIQUE NOT NULL, file_sha256 TEXT NOT NULL, upload_timestamp INTEGER NOT NULL, tags TEXT, local_path TEXT, cache_timestamp INT, link TEXT) STRICT", tableName))
	return err
}

func addMetadataToTable(metadata *FileMetadataRecord, tableName string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (file_name, object_key, file_sha256, upload_timestamp, tags, local_path, cache_timestamp, link) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tableName))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(metadata.FileName, metadata.ObjectKey, metadata.Sha256, metadata.UploadTimestamp, metadata.Tags, metadata.LocalPath, metadata.CacheTimestamp, metadata.Link)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func updateCacheDetailsInTable(metadata *FileMetadataRecord, tableName string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET local_path = ?, cache_timestamp = ? WHERE file_name = ?", tableName))
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

func listFilesInTable(tableName string) ([]FileMetadataRecord, error) {
	rows, err := DB.Query(fmt.Sprintf("SELECT file_name, object_key, file_sha256, upload_timestamp, tags, local_path, cache_timestamp, link FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileMetadataRecord
	for rows.Next() {
		var file FileMetadataRecord
		err = rows.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags, &file.LocalPath, &file.CacheTimestamp, &file.Link)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func getFileMetadataFromTable(filename string, tableName string) (*FileMetadataRecord, error) {
	row := DB.QueryRow(fmt.Sprintf("SELECT file_name, object_key, file_sha256, upload_timestamp, tags, cache_timestamp, link FROM %s WHERE file_name = ?", tableName), filename)
	var file FileMetadataRecord
	err := row.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags, &file.CacheTimestamp, &file.Link)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func queryTags(tagName string, tableName string) ([]FileMetadataRecord, error) {
	rows, err := DB.Query(fmt.Sprintf("SELECT file_name, object_key, file_sha256, upload_timestamp, tags, link FROM %s WHERE tags LIKE ?", tableName), "%"+tagName+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileMetadataRecord
	for rows.Next() {
		var file FileMetadataRecord
		err = rows.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags, &file.Link)
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
