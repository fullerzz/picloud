package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func uploadFileToS3(metadata *FileMetadata, fileContent []byte) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}
	client := s3.NewFromConfig(cfg)
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("my-bucket"), // TODO: Load name of bucket from environment variable
		Key:    aws.String(metadata.Name),
		Body:   bytes.NewReader(fileContent),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://my-bucket.s3.amazonaws.com/%s", metadata.Name), nil
}
