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
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3FilesBucket struct {
	S3Client   *s3.Client
	BucketName string
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
	setupS3(cfg)
}

func setupS3(cfg aws.Config) {
	client := s3.NewFromConfig(cfg)
	bucket := os.Getenv("S3_BUCKET")
	s3Bucket := &S3FilesBucket{S3Client: client, BucketName: bucket}
	filesBucket = *s3Bucket
}
