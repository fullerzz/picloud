package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

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
	Tags            string `dynamodbav:"tags"`
}

type MetadataTable struct {
	DynamoDBClient *dynamodb.Client
	TableName      string
}

func (table *MetadataTable) Query(filename string) ([]MetadataTableItem, error) {
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

func (table *MetadataTable) QueryTags(tagName string) ([]MetadataTableItem, error) {
	// FIXME: This is not working
	var items []MetadataTableItem
	filtEx := expression.Name("tags").Contains(expression.Value(tagName))
	projEx := expression.NamesList(expression.Name("file_name"), expression.Name("object_key"), expression.Name("upload_timestamp"))
	expr, err := expression.NewBuilder().WithFilter(filtEx).WithProjection(projEx).Build()
	if err != nil {
		return nil, err
	}
	scanPaginator := dynamodb.NewScanPaginator(table.DynamoDBClient, &dynamodb.ScanInput{
		TableName:                 aws.String(table.TableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	for scanPaginator.HasMorePages() {
		response, err := scanPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		var itemsPage []MetadataTableItem
		err = attributevalue.UnmarshalListOfMaps(response.Items, &itemsPage)
		if err != nil {
			return nil, err
		}
		items = append(items, itemsPage...)
	}
	return items, nil
}

func (table *MetadataTable) Scan() ([]MetadataTableItem, error) {
	var items []MetadataTableItem
	scanPaginator := dynamodb.NewScanPaginator(table.DynamoDBClient, &dynamodb.ScanInput{
		TableName: aws.String(table.TableName),
	})
	for scanPaginator.HasMorePages() {
		response, err := scanPaginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		var itemsPage []MetadataTableItem
		err = attributevalue.UnmarshalListOfMaps(response.Items, &itemsPage)
		if err != nil {
			return nil, err
		}
		items = append(items, itemsPage...)
	}
	return items, nil
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
