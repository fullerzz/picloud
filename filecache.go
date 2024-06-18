package main

import (
	"io"
	"os"
)

func isStoredLocally(filemetadata *FileMetadataRecord) bool {
	isLocal := filemetadata.LocalPath != nil
	isUpToDate := filemetadata.CacheTimestamp > 0 && filemetadata.CacheTimestamp > filemetadata.UploadTimestamp
	return isLocal && isUpToDate
}

func getLocalFile(filemetadata *FileMetadataRecord) ([]byte, error) {
	file, err := os.Open(*filemetadata.LocalPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func saveFileLocally(filemetadata *FileMetadataRecord, content []byte) error {
	localPath := conf.FilePrefix + filemetadata.FileName
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return err
	}

	filemetadata.CacheTimestamp = getTimestamp()
	filemetadata.LocalPath = strPtr(file.Name())
	err = updateCacheDetailsInTable(filemetadata)
	return err
}

func strPtr(s string) *string {
	return &s
}
