package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

const (
	captchaCodeLength = 5
	captchaTTL        = 5 * time.Minute
)

var captchaAlphabet = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

// ImageCaptchaPayload describes a generated image captcha.
type ImageCaptchaPayload struct {
	CaptchaID string `json:"captchaId"`
	ImageData string `json:"imageData"`
	ExpiresIn int64  `json:"expiresIn"`
}

// ImageCaptchaService generates and verifies image captchas.
type ImageCaptchaService interface {
	Create(ctx context.Context) (ImageCaptchaPayload, error)
	Verify(ctx context.Context, captchaID string, code string) error
}

type imageCaptchaService struct {
	cache repository.CacheStore
	rng   *rand.Rand
}

type captchaRecord struct {
	Code string `json:"code"`
}

// NewImageCaptchaService creates the image captcha service.
func NewImageCaptchaService(cache repository.CacheStore) ImageCaptchaService {
	return &imageCaptchaService{
		cache: cache,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *imageCaptchaService) Create(ctx context.Context) (ImageCaptchaPayload, error) {
	if s.cache == nil {
		return ImageCaptchaPayload{}, fmt.Errorf("captcha cache is not configured")
	}

	code := s.randomCode()
	captchaID := "captcha_" + strings.ReplaceAll(utils.NewRequestID(), "-", "")
	if err := s.cache.SetJSON(ctx, captchaCacheKey(captchaID), captchaRecord{Code: code}, captchaTTL); err != nil {
		return ImageCaptchaPayload{}, err
	}

	return ImageCaptchaPayload{
		CaptchaID: captchaID,
		ImageData: renderCaptchaImage(code, s.rng),
		ExpiresIn: int64(captchaTTL.Seconds()),
	}, nil
}

func (s *imageCaptchaService) Verify(ctx context.Context, captchaID string, code string) error {
	if s.cache == nil {
		return fmt.Errorf("captcha cache is not configured")
	}

	captchaID = strings.TrimSpace(captchaID)
	code = strings.ToUpper(strings.TrimSpace(code))
	if captchaID == "" || code == "" {
		return utils.NewAppError(400, 400, "image captcha is required")
	}

	var record captchaRecord
	if err := s.cache.GetJSON(ctx, captchaCacheKey(captchaID), &record); err != nil {
		if err == repository.ErrCacheMiss {
			return utils.NewAppError(400, 400, "image captcha has expired")
		}
		return err
	}

	if strings.ToUpper(strings.TrimSpace(record.Code)) != code {
		return utils.NewAppError(400, 400, "image captcha is incorrect")
	}

	return nil
}

func (s *imageCaptchaService) randomCode() string {
	buffer := make([]rune, captchaCodeLength)
	for i := range buffer {
		buffer[i] = captchaAlphabet[s.rng.Intn(len(captchaAlphabet))]
	}
	return string(buffer)
}

func renderCaptchaImage(code string, rng *rand.Rand) string {
	lines := make([]string, 0, 4)
	for idx := 0; idx < 4; idx++ {
		lines = append(lines, fmt.Sprintf(
			`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="rgba(37,99,235,0.28)" stroke-width="1.6"/>`,
			rng.Intn(160), rng.Intn(56), rng.Intn(160), rng.Intn(56),
		))
	}

	chars := make([]string, 0, len(code))
	for idx, item := range code {
		rotation := rng.Intn(18) - 9
		y := 34 + rng.Intn(8)
		x := 18 + idx*24
		chars = append(chars, fmt.Sprintf(
			`<text x="%d" y="%d" font-size="28" font-family="Arial, sans-serif" font-weight="700" fill="#1f2937" transform="rotate(%d %d %d)">%c</text>`,
			x, y, rotation, x, y, item,
		))
	}

	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="148" height="52" viewBox="0 0 148 52">
<rect width="148" height="52" rx="12" fill="#f8fbff"/>
<rect x="1" y="1" width="146" height="50" rx="11" fill="none" stroke="#d7e3f4"/>
%s
%s
</svg>`,
		strings.Join(lines, ""),
		strings.Join(chars, ""),
	)

	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

func captchaCacheKey(captchaID string) string {
	return "captcha:" + strings.TrimSpace(captchaID)
}
