package repository

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioRepository interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

type minioRepo struct {
	client *minio.Client
	bucket string
	log    *slog.Logger
}

func NewMinioRepo(endpoint, accessKey, secretKey, bucket string, useSSL bool, log *slog.Logger) (MinioRepository, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
		log.Info("bucket created", "bucket", bucket)
	}

	return &minioRepo{client: client, bucket: bucket, log: log}, nil
}

func (r *minioRepo) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := r.client.PutObject(ctx, r.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (r *minioRepo) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := r.client.PresignedGetObject(ctx, r.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}
	return u.String(), nil
}

func (r *minioRepo) Delete(ctx context.Context, key string) error {
	return r.client.RemoveObject(ctx, r.bucket, key, minio.RemoveObjectOptions{})
}