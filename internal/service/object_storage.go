package service

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	oss "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscredentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/gee-coder/template-go-backend/internal/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"github.com/minio/minio-go/v7"
	miniocredentials "github.com/minio/minio-go/v7/pkg/credentials"
)

// UploadObjectInput describes a single uploaded object.
type UploadObjectInput struct {
	Directory   string
	Filename    string
	Reader      io.Reader
	Size        int64
	ContentType string
}

// StoredObject describes a stored object and its public URL.
type StoredObject struct {
	Key string
	URL string
}

// ObjectStorage stores uploaded files in a configurable backend.
type ObjectStorage interface {
	Upload(ctx context.Context, input UploadObjectInput) (StoredObject, error)
	SupportsPublicURL(rawURL string) bool
}

type localObjectStorage struct {
	rootDir       string
	publicBaseURL string
}

type minioObjectStorage struct {
	client        *minio.Client
	bucket        string
	rootPath      string
	publicBaseURL string
}

type aliyunOSSObjectStorage struct {
	client        *oss.Client
	bucket        string
	rootPath      string
	publicBaseURL string
}

type huaweiOBSObjectStorage struct {
	client        *obs.ObsClient
	bucket        string
	rootPath      string
	publicBaseURL string
}

// NewObjectStorage creates an object storage provider from config.
func NewObjectStorage(cfg *config.Config) (ObjectStorage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage config is required")
	}

	switch cfg.Storage.ResolvedProvider() {
	case "local":
		return &localObjectStorage{
			rootDir:       cfg.App.UploadPath(),
			publicBaseURL: normalizeBaseURL(firstNonEmpty(cfg.Storage.Local.PublicBaseURL, "/uploads")),
		}, nil
	case "minio":
		client, err := minio.New(cfg.Storage.MinIO.Endpoint, &minio.Options{
			Creds:  miniocredentials.NewStaticV4(cfg.Storage.MinIO.AccessKeyID, cfg.Storage.MinIO.AccessKeySecret, ""),
			Secure: cfg.Storage.MinIO.UseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("create minio client: %w", err)
		}
		return &minioObjectStorage{
			client:        client,
			bucket:        strings.TrimSpace(cfg.Storage.MinIO.Bucket),
			rootPath:      normalizeObjectPrefix(cfg.Storage.MinIO.RootPath),
			publicBaseURL: normalizeBaseURL(firstNonEmpty(cfg.Storage.MinIO.PublicBaseURL, buildBucketPublicBaseURL(cfg.Storage.MinIO.UseSSL, cfg.Storage.MinIO.Endpoint, cfg.Storage.MinIO.Bucket, false))),
		}, validateBucketConfig("minio", cfg.Storage.MinIO.Bucket)
	case "aliyun_oss", "aliyun":
		client := oss.NewClient(oss.LoadDefaultConfig().
			WithRegion(strings.TrimSpace(cfg.Storage.AliyunOSS.Region)).
			WithCredentialsProvider(osscredentials.NewStaticCredentialsProvider(cfg.Storage.AliyunOSS.AccessKeyID, cfg.Storage.AliyunOSS.AccessKeySecret, "")).
			WithEndpoint(strings.TrimSpace(cfg.Storage.AliyunOSS.Endpoint)))
		return &aliyunOSSObjectStorage{
			client:        client,
			bucket:        strings.TrimSpace(cfg.Storage.AliyunOSS.Bucket),
			rootPath:      normalizeObjectPrefix(cfg.Storage.AliyunOSS.RootPath),
			publicBaseURL: normalizeBaseURL(firstNonEmpty(cfg.Storage.AliyunOSS.PublicBaseURL, buildBucketPublicBaseURL(true, cfg.Storage.AliyunOSS.Endpoint, cfg.Storage.AliyunOSS.Bucket, true))),
		}, validateBucketConfig("aliyun oss", cfg.Storage.AliyunOSS.Bucket)
	case "huawei_obs", "huawei":
		client, err := obs.New(cfg.Storage.HuaweiOBS.AccessKeyID, cfg.Storage.HuaweiOBS.AccessKeySecret, cfg.Storage.HuaweiOBS.Endpoint)
		if err != nil {
			return nil, fmt.Errorf("create huawei obs client: %w", err)
		}
		return &huaweiOBSObjectStorage{
			client:        client,
			bucket:        strings.TrimSpace(cfg.Storage.HuaweiOBS.Bucket),
			rootPath:      normalizeObjectPrefix(cfg.Storage.HuaweiOBS.RootPath),
			publicBaseURL: normalizeBaseURL(firstNonEmpty(cfg.Storage.HuaweiOBS.PublicBaseURL, buildBucketPublicBaseURL(strings.HasPrefix(strings.ToLower(strings.TrimSpace(cfg.Storage.HuaweiOBS.Endpoint)), "https://"), cfg.Storage.HuaweiOBS.Endpoint, cfg.Storage.HuaweiOBS.Bucket, true))),
		}, validateBucketConfig("huawei obs", cfg.Storage.HuaweiOBS.Bucket)
	default:
		return nil, fmt.Errorf("unsupported storage provider %q", cfg.Storage.ResolvedProvider())
	}
}

func (s *localObjectStorage) Upload(_ context.Context, input UploadObjectInput) (StoredObject, error) {
	objectKey := buildObjectKey("", input.Directory, input.Filename)
	targetPath := filepath.Join(s.rootDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return StoredObject{}, err
	}

	target, err := os.Create(targetPath)
	if err != nil {
		return StoredObject{}, err
	}
	defer target.Close()

	if _, err := io.Copy(target, input.Reader); err != nil {
		return StoredObject{}, err
	}

	return StoredObject{
		Key: objectKey,
		URL: joinPublicURL(s.publicBaseURL, objectKey),
	}, nil
}

func (s *localObjectStorage) SupportsPublicURL(rawURL string) bool {
	return hasPublicURLPrefix(rawURL, s.publicBaseURL)
}

func (s *minioObjectStorage) Upload(ctx context.Context, input UploadObjectInput) (StoredObject, error) {
	objectKey := buildObjectKey(s.rootPath, input.Directory, input.Filename)
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, input.Reader, input.Size, minio.PutObjectOptions{
		ContentType: strings.TrimSpace(input.ContentType),
	})
	if err != nil {
		return StoredObject{}, err
	}

	return StoredObject{
		Key: objectKey,
		URL: joinPublicURL(s.publicBaseURL, objectKey),
	}, nil
}

func (s *minioObjectStorage) SupportsPublicURL(rawURL string) bool {
	return hasPublicURLPrefix(rawURL, s.publicBaseURL)
}

func (s *aliyunOSSObjectStorage) Upload(ctx context.Context, input UploadObjectInput) (StoredObject, error) {
	objectKey := buildObjectKey(s.rootPath, input.Directory, input.Filename)
	request := &oss.PutObjectRequest{
		Bucket:      oss.Ptr(s.bucket),
		Key:         oss.Ptr(objectKey),
		Body:        input.Reader,
		ContentType: oss.Ptr(strings.TrimSpace(input.ContentType)),
	}
	if _, err := s.client.PutObject(ctx, request); err != nil {
		return StoredObject{}, err
	}

	return StoredObject{
		Key: objectKey,
		URL: joinPublicURL(s.publicBaseURL, objectKey),
	}, nil
}

func (s *aliyunOSSObjectStorage) SupportsPublicURL(rawURL string) bool {
	return hasPublicURLPrefix(rawURL, s.publicBaseURL)
}

func (s *huaweiOBSObjectStorage) Upload(_ context.Context, input UploadObjectInput) (StoredObject, error) {
	objectKey := buildObjectKey(s.rootPath, input.Directory, input.Filename)
	request := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			ObjectOperationInput: obs.ObjectOperationInput{
				Bucket: s.bucket,
				Key:    objectKey,
			},
			HttpHeader: obs.HttpHeader{
				ContentType: strings.TrimSpace(input.ContentType),
			},
		},
		Body: input.Reader,
	}
	if input.Size > 0 {
		request.ContentLength = input.Size
	}
	if _, err := s.client.PutObject(request); err != nil {
		return StoredObject{}, err
	}

	return StoredObject{
		Key: objectKey,
		URL: joinPublicURL(s.publicBaseURL, objectKey),
	}, nil
}

func (s *huaweiOBSObjectStorage) SupportsPublicURL(rawURL string) bool {
	return hasPublicURLPrefix(rawURL, s.publicBaseURL)
}

func validateBucketConfig(provider string, bucket string) error {
	if strings.TrimSpace(bucket) == "" {
		return fmt.Errorf("%s bucket is required", provider)
	}
	return nil
}

func buildObjectKey(rootPath string, directory string, filename string) string {
	return strings.TrimLeft(path.Join(
		normalizeObjectPrefix(rootPath),
		strings.Trim(strings.TrimSpace(directory), "/"),
		strings.Trim(strings.TrimSpace(filename), "/"),
	), "/")
}

func normalizeObjectPrefix(value string) string {
	return strings.Trim(strings.TrimSpace(value), "/")
}

func joinPublicURL(baseURL string, objectKey string) string {
	baseURL = normalizeBaseURL(baseURL)
	if baseURL == "" {
		return "/" + strings.TrimLeft(objectKey, "/")
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

func normalizeBaseURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasSuffix(value, "/") {
		return strings.TrimRight(value, "/")
	}
	return value
}

func hasPublicURLPrefix(rawURL string, baseURL string) bool {
	rawURL = strings.TrimSpace(rawURL)
	baseURL = normalizeBaseURL(baseURL)
	if rawURL == "" || baseURL == "" {
		return false
	}

	if strings.HasPrefix(baseURL, "/") {
		return strings.HasPrefix(rawURL, baseURL+"/")
	}
	return strings.HasPrefix(rawURL, baseURL+"/")
}

func buildBucketPublicBaseURL(useSSL bool, endpoint string, bucket string, bucketOnHost bool) string {
	endpoint = strings.TrimSpace(endpoint)
	bucket = strings.TrimSpace(bucket)
	if endpoint == "" || bucket == "" {
		return ""
	}

	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	if strings.HasPrefix(strings.ToLower(endpoint), "http://") || strings.HasPrefix(strings.ToLower(endpoint), "https://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return ""
		}
		host := parsed.Host
		if host == "" {
			host = parsed.Path
		}
		if bucketOnHost {
			return fmt.Sprintf("%s://%s.%s", parsed.Scheme, bucket, host)
		}
		return fmt.Sprintf("%s://%s/%s", parsed.Scheme, host, bucket)
	}
	if bucketOnHost {
		return fmt.Sprintf("%s://%s.%s", scheme, bucket, endpoint)
	}
	return fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucket)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isTrustedUploadedAssetURL(value string, validator func(string) bool) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if validator != nil && validator(value) {
		return true
	}
	if strings.HasPrefix(value, "/uploads/avatars/") {
		return true
	}
	if parsed, err := url.Parse(value); err == nil {
		return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && parsed.Path != ""
	}
	return false
}
