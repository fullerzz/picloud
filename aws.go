package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	b64 "encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3FilesBucket struct {
	S3Client   *s3.Client
	BucketName string
}

type MetadataTableItem struct {
	FileName        string `dynamodbav:"file_name"`
	ObjectKey       string `dynamodbav:"object_key"`
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

func (table *MetadataTable) Query(filename string) ([]MetadataTableItem, error) {
	slog.Debug("Querying metadata table", "filename", filename)
	fmt.Printf("Querying metadata table for %s\n", filename)
	var response *dynamodb.QueryOutput
	var items []MetadataTableItem
	keyExp := expression.Key("file_name").Equal(expression.Value(filename))
	expr, err := expression.NewBuilder().WithKeyCondition(keyExp).Build()
	if err != nil {
		return nil, err
	}
	queryPaginator := dynamodb.NewQueryPaginator(table.DynamoDBClient, &dynamodb.QueryInput{
		TableName:                 aws.String(table.TableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	for queryPaginator.HasMorePages() {
		response, err = queryPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		var itemsPage []MetadataTableItem
		err = attributevalue.UnmarshalListOfMaps(response.Items, &itemsPage)
		if err != nil {
			slog.Error("Error unmarshalling items", "err", err)
			return nil, err
		}
		items = append(items, itemsPage...)
	}
	return items, err
}

func writeMetadataToTable(file *FileUpload, objectKey string) error {
	metadata := &MetadataTableItem{FileName: file.Name, ObjectKey: objectKey, Sha256: getSha256Checksum(&file.Content), FileExtension: "TODO", UploadTimestamp: getTimestamp()}
	return metadataTable.addMetadata(metadata)
}

func getObjectKey(filename string) (string, error) {
	items, err := metadataTable.Query(filename)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("File not found in metadata table")
	}
	// TODO: handle multiple files with the same name
	return items[0].ObjectKey, nil

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

func (bucket *S3FilesBucket) UploadFile(file *FileUpload) (string, error) {
	key := file.Name
	_, err := bucket.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket.BucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(file.Content),
	})
	return key, err
}

func (bucket *S3FilesBucket) DownloadFile(key string) ([]byte, error) {
	result, err := bucket.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Printf("Error downloading file: %s\n", err)
		return nil, err
	}
	defer result.Body.Close()
	fileContents, err := io.ReadAll(result.Body)
	return fileContents, err
}

func createClientConnections() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("Error loading default config", "err", err)
		panic(err)
	}
	setupDynamo(cfg)
	setupS3(cfg)
}

func setupDynamo(cfg aws.Config) {
	client := dynamodb.NewFromConfig(cfg)
	table := &MetadataTable{DynamoDBClient: client, TableName: os.Getenv("DYNAMODB_TABLE")}
	metadataTable = *table
}

func setupS3(cfg aws.Config) {
	client := s3.NewFromConfig(cfg)
	bucket := os.Getenv("S3_BUCKET")
	s3Bucket := &S3FilesBucket{S3Client: client, BucketName: bucket}
	filesBucket = *s3Bucket
}
