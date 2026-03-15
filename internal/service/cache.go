package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gee-coder/template-go-backend/internal/repository"
)

const (
	authOptionsCacheKey      = "auth:options"
	brandingSettingsCacheKey = "branding:settings"
	authOptionsCacheTTL      = 10 * time.Minute
	brandingCacheTTL         = 10 * time.Minute
	permissionCacheTTL       = 5 * time.Minute
)

func permissionCacheKey(userID uint) string {
	return fmt.Sprintf("permissions:user:%d", userID)
}

func invalidatePermissionCache(ctx context.Context, cache repository.CacheStore, userID uint) {
	if cache == nil {
		return
	}
	_ = cache.Delete(ctx, permissionCacheKey(userID))
}

func invalidateAllPermissionCaches(ctx context.Context, cache repository.CacheStore) {
	if cache == nil {
		return
	}
	_ = cache.DeleteByPrefix(ctx, "permissions:user:")
}
