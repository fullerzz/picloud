package main

import (
	"bytes"
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

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
