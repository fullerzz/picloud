package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

// ...

type mockPutObjectAPI func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

func (m mockPutObjectAPI) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m(ctx, params, optFns...)
}

func TestUploadObjectToS3(t *testing.T) {
	var mockFileUpload *FileUpload

	// load baxter.jpg into memory
	testFilePath := "testdata/baxter.jpg"
	fileData, err := os.ReadFile(testFilePath)
	if err != nil {
		fmt.Printf("Error reading file %s\n", testFilePath)
		panic(err)
	}

	mockFileUpload = &FileUpload{
		Name:    "barKey",
		Size:    len(fileData),
		Content: fileData,
		Tags:    "baxter, dog",
	}
	cases := []struct {
		client func(t *testing.T) S3PutObjectAPI
		bucket string
		key    string
		expect *string
	}{
		{
			client: func(t *testing.T) S3PutObjectAPI {
				return mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					t.Helper()
					if params.Bucket == nil {
						t.Fatal("expect bucket to not be nil")
					}
					if e, a := "TestBucket", *params.Bucket; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}
					if params.Key == nil {
						t.Fatal("expect key to not be nil")
					}
					if e, a := mockFileUpload.Name, *params.Key; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return &s3.PutObjectOutput{}, nil
				})
			},
			bucket: "TestBucket",
			key:    mockFileUpload.Name,
			expect: nil,
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			objectKey, err := UploadObject(mockFileUpload, tt.client(t), tt.bucket)
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}
			assert.Equal(t, tt.key, objectKey)
		})
	}
}
