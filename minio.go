package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	Client     *minio.Client
	BucketName string
	Endpoint   string
}

var GlobalMinio *MinioStorage

func NewMinioStorage() (*MinioStorage, error) {
	endpoint := os.Getenv("S3_URL")
	accessKeyID := os.Getenv("S3_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("S3_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("S3_BUCKET")
	useSSL := true

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinioStorage{
		Client:     minioClient,
		BucketName: bucketName,
		Endpoint:   endpoint,
	}, nil
}

func (m *MinioStorage) UploadFile(ctx context.Context, objectName string, file io.Reader, fileSize int64, contentType string) (string, error) {
	_, err := m.Client.PutObject(ctx, m.BucketName, objectName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://%s/%s/%s", m.Endpoint, m.BucketName, objectName)
	return url, nil
}

func (m *MinioStorage) DownloadFile(ctx context.Context, objectName string) (io.Reader, error) {
	object, err := m.Client.GetObject(ctx, m.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return object, nil
}

func InitGlobalMinio() error {
	minioInstance, err := NewMinioStorage()
	if err != nil {
		return err
	}
	GlobalMinio = minioInstance
	return nil
}

func DeleteFileFromMinio(objectPath string) error {
	if GlobalMinio == nil {
		return fmt.Errorf("Minio client not initialized")
	}
	return GlobalMinio.DeleteFile(context.Background(), objectPath)
}

func (m *MinioStorage) DeleteFile(ctx context.Context, objectName string) error {
	err := m.Client.RemoveObject(ctx, m.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (m *MinioStorage) CleanOldBackups(ctx context.Context, retentionDays int, prefix string) error {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	for object := range m.Client.ListObjects(ctx, m.BucketName, opts) {
		if object.Err != nil {
			fmt.Println("Error listing object:", object.Err)
			continue
		}

		if object.LastModified.Before(cutoffDate) {
			fmt.Printf("Deleting old backup: %s (Last modified: %s)\n", object.Key, object.LastModified)
			err := m.DeleteFile(ctx, object.Key)
			if err != nil {
				fmt.Println("Failed to delete object:", err)
			}
		}
	}
	return nil
}
