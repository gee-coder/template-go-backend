package service

import (
	"hash/crc32"
	"net/http"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

var supportedAvatarKeys = []string{
	"default-01",
	"default-02",
	"default-03",
	"default-04",
	"default-05",
	"default-06",
	"default-07",
	"default-08",
}

func normalizeAvatarChoice(value string, uploadedURLValidator func(string) bool, fallbackSeeds ...string) (string, error) {
	value = strings.TrimSpace(value)
	if value != "" {
		if !isSupportedAvatarKey(value) && !isTrustedUploadedAssetURL(value, uploadedURLValidator) {
			return "", utils.NewAppError(http.StatusBadRequest, http.StatusBadRequest, "unsupported avatar option")
		}
		return value, nil
	}

	return defaultAvatarForSeeds(fallbackSeeds...), nil
}

func resolveUserAvatar(user *model.User) string {
	if user == nil {
		return supportedAvatarKeys[0]
	}

	if isSupportedAvatarKey(user.Avatar) {
		return user.Avatar
	}
	if isTrustedUploadedAssetURL(user.Avatar, nil) {
		return strings.TrimSpace(user.Avatar)
	}

	return defaultAvatarForSeeds(user.Username, user.Email, user.Phone, user.Nickname)
}

func defaultAvatarForSeeds(seeds ...string) string {
	for _, seed := range seeds {
		seed = strings.TrimSpace(seed)
		if seed == "" {
			continue
		}
		index := crc32.ChecksumIEEE([]byte(seed)) % uint32(len(supportedAvatarKeys))
		return supportedAvatarKeys[index]
	}

	return supportedAvatarKeys[0]
}

func isSupportedAvatarKey(value string) bool {
	for _, item := range supportedAvatarKeys {
		if item == value {
			return true
		}
	}
	return false
}
