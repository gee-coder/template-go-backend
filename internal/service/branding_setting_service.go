package service

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

var brandingColorPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// BrandingTheme describes brand theme colors returned to clients.
type BrandingTheme struct {
	Primary     string `json:"primary"`
	PrimaryDark string `json:"primaryDark"`
	ShellStart  string `json:"shellStart"`
	ShellEnd    string `json:"shellEnd"`
	HeroStart   string `json:"heroStart"`
	HeroEnd     string `json:"heroEnd"`
}

// BrandingSettings describes the runtime brand config exposed to admin and public pages.
type BrandingSettings struct {
	AppTitle       string        `json:"appTitle"`
	ConsoleName    string        `json:"consoleName"`
	ProductTagline string        `json:"productTagline"`
	LogoMarkURL    string        `json:"logoMarkUrl"`
	FaviconURL     string        `json:"faviconUrl"`
	LoginHeroURL   string        `json:"loginHeroUrl"`
	Theme          BrandingTheme `json:"theme"`
}

// UpdateBrandingSettingInput describes admin brand updates.
type UpdateBrandingSettingInput = BrandingSettings

// UploadBrandingAssetInput describes an uploaded brand asset.
type UploadBrandingAssetInput struct {
	Kind string
	File *multipart.FileHeader
}

// BrandingAssetPayload describes the saved asset location returned to admin.
type BrandingAssetPayload struct {
	URL string `json:"url"`
}

// BrandingSettingService provides admin-facing branding operations.
type BrandingSettingService interface {
	Get(ctx context.Context) (BrandingSettings, error)
	Update(ctx context.Context, input UpdateBrandingSettingInput) (BrandingSettings, error)
	UploadAsset(ctx context.Context, input UploadBrandingAssetInput) (BrandingAssetPayload, error)
}

type brandingSettingService struct {
	repo    repository.BrandingSettingRepository
	storage ObjectStorage
	cache   repository.CacheStore
}

// NewBrandingSettingService creates a branding setting service.
func NewBrandingSettingService(repo repository.BrandingSettingRepository, storage ObjectStorage, cache repository.CacheStore) BrandingSettingService {
	return &brandingSettingService{
		repo:    repo,
		storage: storage,
		cache:   cache,
	}
}

func (s *brandingSettingService) Get(ctx context.Context) (BrandingSettings, error) {
	if s.cache != nil {
		var cached BrandingSettings
		if err := s.cache.GetJSON(ctx, brandingSettingsCacheKey, &cached); err == nil {
			return cached, nil
		} else if err != nil && err != repository.ErrCacheMiss {
			return BrandingSettings{}, err
		}
	}

	setting, err := s.repo.Get(ctx)
	if err != nil {
		if err == utils.ErrNotFound {
			defaults := defaultBrandingSettings()
			if s.cache != nil {
				_ = s.cache.SetJSON(ctx, brandingSettingsCacheKey, defaults, brandingCacheTTL)
			}
			return defaults, nil
		}
		return BrandingSettings{}, err
	}

	result := brandingSettingsFromModel(setting)
	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, brandingSettingsCacheKey, result, brandingCacheTTL)
	}
	return result, nil
}

func (s *brandingSettingService) Update(ctx context.Context, input UpdateBrandingSettingInput) (BrandingSettings, error) {
	setting, err := s.repo.Get(ctx)
	if err != nil {
		if err != utils.ErrNotFound {
			return BrandingSettings{}, err
		}
		setting = &model.BrandingSetting{}
	}

	next := normalizeBrandingSettings(input)
	applyBrandingSettings(setting, next)

	if err := s.repo.Save(ctx, setting); err != nil {
		return BrandingSettings{}, err
	}

	result := brandingSettingsFromModel(setting)
	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, brandingSettingsCacheKey, result, brandingCacheTTL)
	}
	return result, nil
}

func (s *brandingSettingService) UploadAsset(ctx context.Context, input UploadBrandingAssetInput) (BrandingAssetPayload, error) {
	if input.File == nil {
		return BrandingAssetPayload{}, utils.NewAppError(400, 400, "please choose an image")
	}
	if input.File.Size > 5*1024*1024 {
		return BrandingAssetPayload{}, utils.NewAppError(400, 400, "branding image size must be 5MB or smaller")
	}

	kind := strings.TrimSpace(input.Kind)
	if kind != "logoMark" && kind != "favicon" && kind != "loginHero" {
		return BrandingAssetPayload{}, utils.NewAppError(400, 400, "unsupported branding asset kind")
	}

	ext := strings.ToLower(filepath.Ext(input.File.Filename))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".svg", ".webp", ".ico":
	default:
		return BrandingAssetPayload{}, utils.NewAppError(400, 400, "branding images must be PNG, JPG, JPEG, SVG, WEBP, or ICO")
	}

	source, err := input.File.Open()
	if err != nil {
		return BrandingAssetPayload{}, err
	}
	defer source.Close()

	stored, err := s.storage.Upload(ctx, UploadObjectInput{
		Directory:   "branding",
		Filename:    kind + "-" + strings.ReplaceAll(utils.NewRequestID(), "-", "") + ext,
		Reader:      source,
		Size:        input.File.Size,
		ContentType: input.File.Header.Get("Content-Type"),
	})
	if err != nil {
		return BrandingAssetPayload{}, err
	}

	return BrandingAssetPayload{URL: stored.URL}, nil
}

func defaultBrandingSettings() BrandingSettings {
	return BrandingSettings{
		AppTitle:       "Nex Console",
		ConsoleName:    "Nex Console",
		ProductTagline: "A configurable admin console template for business systems",
		LogoMarkURL:    "/branding/logo-mark.svg",
		FaviconURL:     "/branding/logo-mark.svg",
		LoginHeroURL:   "/branding/login-hero.svg",
		Theme: BrandingTheme{
			Primary:     "#2563eb",
			PrimaryDark: "#1d4ed8",
			ShellStart:  "#f5f7fc",
			ShellEnd:    "#eaf0fb",
			HeroStart:   "#2f63f6",
			HeroEnd:     "#1946bd",
		},
	}
}

func brandingSettingsFromModel(setting *model.BrandingSetting) BrandingSettings {
	if setting == nil {
		return defaultBrandingSettings()
	}

	return normalizeBrandingSettings(BrandingSettings{
		AppTitle:       setting.AppTitle,
		ConsoleName:    setting.ConsoleName,
		ProductTagline: setting.ProductTagline,
		LogoMarkURL:    setting.LogoMarkURL,
		FaviconURL:     setting.FaviconURL,
		LoginHeroURL:   setting.LoginHeroURL,
		Theme: BrandingTheme{
			Primary:     setting.Primary,
			PrimaryDark: setting.PrimaryDark,
			ShellStart:  setting.ShellStart,
			ShellEnd:    setting.ShellEnd,
			HeroStart:   setting.HeroStart,
			HeroEnd:     setting.HeroEnd,
		},
	})
}

func applyBrandingSettings(target *model.BrandingSetting, input BrandingSettings) {
	target.AppTitle = input.AppTitle
	target.ConsoleName = input.ConsoleName
	target.ProductTagline = input.ProductTagline
	target.LogoMarkURL = strings.TrimSpace(input.LogoMarkURL)
	target.FaviconURL = strings.TrimSpace(input.FaviconURL)
	target.LoginHeroURL = strings.TrimSpace(input.LoginHeroURL)
	target.Primary = input.Theme.Primary
	target.PrimaryDark = input.Theme.PrimaryDark
	target.ShellStart = input.Theme.ShellStart
	target.ShellEnd = input.Theme.ShellEnd
	target.HeroStart = input.Theme.HeroStart
	target.HeroEnd = input.Theme.HeroEnd
}

func normalizeBrandingSettings(input BrandingSettings) BrandingSettings {
	defaults := defaultBrandingSettings()
	return BrandingSettings{
		AppTitle:       normalizeText(input.AppTitle, defaults.AppTitle),
		ConsoleName:    normalizeText(input.ConsoleName, defaults.ConsoleName),
		ProductTagline: normalizeText(input.ProductTagline, defaults.ProductTagline),
		LogoMarkURL:    strings.TrimSpace(input.LogoMarkURL),
		FaviconURL:     strings.TrimSpace(input.FaviconURL),
		LoginHeroURL:   strings.TrimSpace(input.LoginHeroURL),
		Theme: BrandingTheme{
			Primary:     normalizeHexColor(input.Theme.Primary, defaults.Theme.Primary),
			PrimaryDark: normalizeHexColor(input.Theme.PrimaryDark, defaults.Theme.PrimaryDark),
			ShellStart:  normalizeHexColor(input.Theme.ShellStart, defaults.Theme.ShellStart),
			ShellEnd:    normalizeHexColor(input.Theme.ShellEnd, defaults.Theme.ShellEnd),
			HeroStart:   normalizeHexColor(input.Theme.HeroStart, defaults.Theme.HeroStart),
			HeroEnd:     normalizeHexColor(input.Theme.HeroEnd, defaults.Theme.HeroEnd),
		},
	}
}

func normalizeText(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func normalizeHexColor(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !brandingColorPattern.MatchString(trimmed) {
		return fallback
	}
	return strings.ToLower(trimmed)
}
