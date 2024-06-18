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
	// TODO: Build string from filemetadata.LocalPath and figure out where local files are stored on server
	file, err := os.Create(*filemetadata.LocalPath) // FIXME: I don't think the LocalPath field is populated anywhere yet
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
	err = updateCacheDetails(filemetadata)
	return err
}

func strPtr(s string) *string {
	return &s
}
