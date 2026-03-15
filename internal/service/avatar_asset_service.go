package service

import (
	"context"
	"mime/multipart"
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
	storage ObjectStorage
}

// NewAvatarAssetService creates an avatar upload service.
func NewAvatarAssetService(storage ObjectStorage) AvatarAssetService {
	return &avatarAssetService{storage: storage}
}

func (s *avatarAssetService) Upload(ctx context.Context, input UploadAvatarAssetInput) (AvatarAssetPayload, error) {
	if input.File == nil {
		return AvatarAssetPayload{}, utils.NewAppError(400, 400, "please choose an avatar image")
	}
	if input.File.Size > avatarMaxUploadSize {
		return AvatarAssetPayload{}, utils.NewAppError(400, 400, "avatar image size must be 2MB or smaller")
	}

	ext := strings.ToLower(filepath.Ext(input.File.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		return AvatarAssetPayload{}, utils.NewAppError(400, 400, "avatar images must be PNG, JPG, JPEG, or WEBP")
	}

	source, err := input.File.Open()
	if err != nil {
		return AvatarAssetPayload{}, err
	}
	defer source.Close()

	stored, err := s.storage.Upload(ctx, UploadObjectInput{
		Directory:   "avatars",
		Filename:    "avatar-" + strings.ReplaceAll(utils.NewRequestID(), "-", "") + ext,
		Reader:      source,
		Size:        input.File.Size,
		ContentType: input.File.Header.Get("Content-Type"),
	})
	if err != nil {
		return AvatarAssetPayload{}, err
	}

	return AvatarAssetPayload{URL: stored.URL}, nil
}
