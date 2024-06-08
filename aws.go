package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MetadataTableItem struct {
	FileName        string `dynamodbav:"file_name"`
	Sha256          string `dynamodbav:"file_sha256"`
	FileExtension   string `dynamodbav:"file_extension"`
	UploadTimestamp int64  `dynamodbav:"upload_timestamp"`
}

type MetadataTable struct {
	DynamoDBClient *dynamodb.Client
	TableName      string
}

func (table *MetadataTable) addMetadata(metadata *MetadataTableItem) error {
	item, err := attributevalue.MarshalMap(metadata)
	if err != nil {
		panic(err)
	}
	_, err = table.DynamoDBClient.PutItem(context.TODO(), &dynamodb.PutItemInput{TableName: aws.String(table.TableName), Item: item})
	return err
}

func writeMetadataToTable(fileName string, fileContent []byte) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	client := dynamodb.NewFromConfig(cfg)
	table := &MetadataTable{DynamoDBClient: client, TableName: os.Getenv("DYNAMODB_TABLE")}
	metadata := &MetadataTableItem{FileName: fileName, Sha256: getSha256Checksum(&fileContent), FileExtension: "TODO", UploadTimestamp: getTimestamp()}
	return table.addMetadata(metadata)
}

func getSha256Checksum(fileContent *[]byte) string {
	h := sha256.New()
	_, err := io.Copy(h, bytes.NewReader(*fileContent))
	if err != nil {
		slog.Error("Error calculating copying bytes in getSha256Checksum")
		panic(err)
	}
	checksum := b64.StdEncoding.EncodeToString(h.Sum(nil))
	slog.Debug("Checksum calculated", "checksum", checksum)
	return checksum
}

func getTimestamp() int64 {
	return time.Now().UTC().UnixMilli()
}

func uploadFileToS3(metadata *FileMetadata, fileContent []byte) error {
	bucket := os.Getenv("S3_BUCKET")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(cfg)
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(metadata.Name),
		Body:   bytes.NewReader(fileContent),
	})
	if err != nil {
		return err
	}
	return nil
}
