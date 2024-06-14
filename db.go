package main

import "database/sql"

type SQLTableItem struct {
	FileName        string `json:"file_name"`
	ObjectKey       string `json:"object_key"`
	Sha256          string `json:"file_sha256"`
	UploadTimestamp int64  `json:"upload_timestamp"`
	Tags            string `json:"tags"` // comma separated tags
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

func addMetadataToTable(metadata *SQLTableItem) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO file_metadata (file_name, object_key, file_sha256, upload_timestamp, tags) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(metadata.FileName, metadata.ObjectKey, metadata.Sha256, metadata.UploadTimestamp, metadata.Tags)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func listFilesInTable() ([]SQLTableItem, error) {
	rows, err := DB.Query("SELECT file_name, object_key, file_sha256, upload_timestamp, tags FROM file_metadata")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []SQLTableItem
	for rows.Next() {
		var file SQLTableItem
		err = rows.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func getFileMetadataFromTable(filename string) (*SQLTableItem, error) {
	row := DB.QueryRow("SELECT file_name, object_key, file_sha256, upload_timestamp, tags FROM file_metadata WHERE file_name = ?", filename)
	var file SQLTableItem
	err := row.Scan(&file.FileName, &file.ObjectKey, &file.Sha256, &file.UploadTimestamp, &file.Tags)
	if err != nil {
		return nil, err
	}
	return &file, nil
}
