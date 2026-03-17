package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Storage struct {
	bucket        string
	publicBaseURL string
	presignExpire time.Duration
	client        *s3.Client
	presignClient *s3.PresignClient
}

// NewS3Storage 创建 S3 存储实现。
func NewS3Storage(cfg S3Config) (Backend, error) {
	bucket := strings.TrimSpace(cfg.Bucket)
	if bucket == "" {
		return nil, ErrInvalidS3Bucket
	}

	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		return nil, ErrInvalidS3Region
	}

	accessKey := strings.TrimSpace(cfg.AccessKey)
	if accessKey == "" {
		return nil, ErrInvalidS3AccessKey
	}

	secretKey := strings.TrimSpace(cfg.SecretKey)
	if secretKey == "" {
		return nil, ErrInvalidS3SecretKey
	}

	cfgAWS := aws.Config{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}

	client := s3.NewFromConfig(cfgAWS, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		endpoint := strings.TrimSpace(cfg.Endpoint)
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	expire := cfg.PresignExpire
	if expire <= 0 {
		expire = 15 * time.Minute
	}

	return &s3Storage{
		bucket:        bucket,
		publicBaseURL: strings.TrimSpace(cfg.PublicBaseURL),
		presignExpire: expire,
		client:        client,
		presignClient: s3.NewPresignClient(client),
	}, nil
}

func (s *s3Storage) Upload(ctx context.Context, objectKey string, data []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(data),
	}
	if strings.TrimSpace(contentType) != "" {
		input.ContentType = aws.String(contentType)
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("s3 上传失败: %w", err)
	}
	return nil
}

func (s *s3Storage) Download(ctx context.Context, objectKey string) ([]byte, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 下载失败: %w", err)
	}
	defer out.Body.Close()

	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 s3 对象内容失败: %w", err)
	}
	return data, nil
}

func (s *s3Storage) Delete(ctx context.Context, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("s3 删除失败: %w", err)
	}
	return nil
}

func (s *s3Storage) URL(ctx context.Context, objectKey string) (string, error) {
	if s.publicBaseURL != "" {
		return strings.TrimRight(s.publicBaseURL, "/") + "/" + strings.TrimLeft(objectKey, "/"), nil
	}

	out, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(objectKey),
	}, func(options *s3.PresignOptions) {
		options.Expires = s.presignExpire
	})
	if err != nil {
		return "", fmt.Errorf("生成 s3 访问地址失败: %w", err)
	}

	return out.URL, nil
}
