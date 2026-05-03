package upload

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Uploader uploads raw bytes to object storage and returns the stored path.
type Uploader interface {
	Upload(ctx context.Context, key string, content io.Reader) (path string, err error)
	// PublicURL returns the browser-accessible HTTPS URL for the given key.
	PublicURL(key string) string
}

// S3Config holds credentials for an S3-compatible storage backend (e.g. Supabase Storage).
type S3Config struct {
	Endpoint      string // e.g. https://<project>.storage.supabase.co/storage/v1/s3
	PublicBaseURL string // e.g. https://<project>.supabase.co/storage/v1/object/public/<bucket>
	Region        string
	Bucket        string
	Prefix        string // e.g. "agents"
	AccessKey     string
	SecretKey     string
}

// S3Uploader uploads files to an S3-compatible backend.
type S3Uploader struct {
	cfg    S3Config
	client *s3.Client
}

// NewS3Uploader returns an S3Uploader configured for the given S3-compatible endpoint.
func NewS3Uploader(cfg S3Config) *S3Uploader {
	endpoint := strings.TrimPrefix(cfg.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	s3client := s3.New(s3.Options{
		BaseEndpoint: aws.String("https://" + endpoint),
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: true,
		HTTPClient:   &http.Client{},
	})
	return &S3Uploader{cfg: cfg, client: s3client}
}

// Upload writes content to <bucket>/<prefix>/<key> and returns the S3 path.
func (u *S3Uploader) Upload(ctx context.Context, key string, content io.Reader) (string, error) {
	fullKey := key
	if u.cfg.Prefix != "" {
		fullKey = u.cfg.Prefix + "/" + key
	}

	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.cfg.Bucket),
		Key:    aws.String(fullKey),
		Body:   content,
	})
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", fullKey, err)
	}

	return fmt.Sprintf("s3://%s/%s", u.cfg.Bucket, fullKey), nil
}

// PublicURL returns the browser-accessible HTTPS URL for the given storage key.
func (u *S3Uploader) PublicURL(key string) string {
	fullKey := key
	if u.cfg.Prefix != "" {
		fullKey = u.cfg.Prefix + "/" + key
	}
	base := strings.TrimRight(u.cfg.PublicBaseURL, "/")
	return base + "/" + fullKey
}
