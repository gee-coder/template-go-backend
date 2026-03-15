package service

import (
	"context"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/utils"
)

const (
	avatarMaxUploadSize = 2 * 1024 * 1024
)

// UploadAvatarAssetInput describes an uploaded avatar image.
type UploadAvatarAssetInput struct {
	File *multipart.FileHeader
}

// AvatarAssetPayload describes the saved avatar location.
type AvatarAssetPayload struct {
	URL string `json:"url"`
}

// AvatarAssetService provides avatar upload capabilities.
type AvatarAssetService interface {
	Upload(ctx context.Context, input UploadAvatarAssetInput) (AvatarAssetPayload, error)
}

type avatarAssetService struct {
	uploadDir string
}

// NewAvatarAssetService creates an avatar upload service.
func NewAvatarAssetService(uploadDir string) AvatarAssetService {
	return &avatarAssetService{uploadDir: uploadDir}
}

func (s *avatarAssetService) Upload(_ context.Context, input UploadAvatarAssetInput) (AvatarAssetPayload, error) {
	if input.File == nil {
		return AvatarAssetPayload{}, utils.NewAppError(400, 400, "请选择要上传的头像图片")
	}
	if input.File.Size > avatarMaxUploadSize {
		return AvatarAssetPayload{}, utils.NewAppError(400, 400, "头像图片大小不能超过 2MB")
	}

	ext := strings.ToLower(filepath.Ext(input.File.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		return AvatarAssetPayload{}, utils.NewAppError(400, 400, "头像仅支持 png、jpg、jpeg、webp 格式")
	}

	targetDir := filepath.Join(s.uploadDir, "avatars")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return AvatarAssetPayload{}, err
	}

	filename := "avatar-" + strings.ReplaceAll(utils.NewRequestID(), "-", "") + ext
	targetPath := filepath.Join(targetDir, filename)

	source, err := input.File.Open()
	if err != nil {
		return AvatarAssetPayload{}, err
	}
	defer source.Close()

	target, err := os.Create(targetPath)
	if err != nil {
		return AvatarAssetPayload{}, err
	}
	defer target.Close()

	if _, err := target.ReadFrom(source); err != nil {
		return AvatarAssetPayload{}, err
	}

	return AvatarAssetPayload{URL: "/uploads/avatars/" + filename}, nil
}
