package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

type fakeUserRepository struct {
	users        map[string]*model.User
	usersByID    map[uint]*model.User
	usersByEmail map[string]*model.User
	usersByPhone map[string]*model.User
	permissions  []string
	listResponse []model.User
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	if f.usersByID != nil {
		if user, ok := f.usersByID[id]; ok {
			return user, nil
		}
	}
	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, utils.ErrNotFound
}

func (f *fakeUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, ok := f.users[username]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return user, nil
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if f.usersByEmail != nil {
		if user, ok := f.usersByEmail[email]; ok {
			return user, nil
		}
	}
	for _, user := range f.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, utils.ErrNotFound
}

func (f *fakeUserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	if f.usersByPhone != nil {
		if user, ok := f.usersByPhone[phone]; ok {
			return user, nil
		}
	}
	for _, user := range f.users {
		if user.Phone == phone {
			return user, nil
		}
	}
	return nil, utils.ErrNotFound
}

func (f *fakeUserRepository) List(ctx context.Context, filter repository.UserFilter) ([]model.User, error) {
	if f.listResponse != nil {
		return f.listResponse, nil
	}
	items := make([]model.User, 0, len(f.users))
	for _, user := range f.users {
		items = append(items, *user)
	}
	return items, nil
}

func (f *fakeUserRepository) Create(ctx context.Context, user *model.User) error {
	if f.users == nil {
		f.users = map[string]*model.User{}
	}
	if f.usersByID == nil {
		f.usersByID = map[uint]*model.User{}
	}
	if f.usersByEmail == nil {
		f.usersByEmail = map[string]*model.User{}
	}
	if f.usersByPhone == nil {
		f.usersByPhone = map[string]*model.User{}
	}
	user.ID = uint(len(f.users) + 1)
	f.users[user.Username] = user
	f.usersByID[user.ID] = user
	if user.Email != "" {
		f.usersByEmail[user.Email] = user
	}
	if user.Phone != "" {
		f.usersByPhone[user.Phone] = user
	}
	return nil
}

func (f *fakeUserRepository) Update(ctx context.Context, user *model.User) error {
	if _, ok := f.users[user.Username]; !ok {
		return utils.ErrNotFound
	}
	f.users[user.Username] = user
	if f.usersByID != nil {
		f.usersByID[user.ID] = user
	}
	if f.usersByEmail != nil && user.Email != "" {
		f.usersByEmail[user.Email] = user
	}
	if f.usersByPhone != nil && user.Phone != "" {
		f.usersByPhone[user.Phone] = user
	}
	return nil
}

func (f *fakeUserRepository) Delete(ctx context.Context, id uint) error { return nil }

func (f *fakeUserRepository) ReplaceRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return nil
}

func (f *fakeUserRepository) GetPermissions(ctx context.Context, userID uint) ([]string, error) {
	return f.permissions, nil
}

type fakeTokenStore struct {
	tokens map[string]uint
}

type fakeCacheStore struct {
	values map[string][]byte
}

type fakeSMSVerificationService struct {
	sendErr   error
	sent      []SendSMSCodeInput
	verifyErr error
	verified  []VerifySMSCodeInput
}

type fakeEmailVerificationService struct {
	sendErr   error
	sent      []SendEmailCodeInput
	verifyErr error
	verified  []VerifyEmailCodeInput
}

type fakeImageCaptchaService struct {
	verifyErr error
	verified  [][2]string
}

func (f *fakeTokenStore) Save(ctx context.Context, refreshToken string, userID uint, ttl time.Duration) error {
	f.tokens[refreshToken] = userID
	return nil
}

func (f *fakeTokenStore) Get(ctx context.Context, refreshToken string) (uint, error) {
	userID, ok := f.tokens[refreshToken]
	if !ok {
		return 0, errors.New("missing refresh token")
	}
	return userID, nil
}

func (f *fakeTokenStore) Delete(ctx context.Context, refreshToken string) error {
	delete(f.tokens, refreshToken)
	return nil
}

func (f *fakeCacheStore) GetJSON(ctx context.Context, key string, target any) error {
	if f.values == nil {
		return repository.ErrCacheMiss
	}
	payload, ok := f.values[key]
	if !ok {
		return repository.ErrCacheMiss
	}
	return json.Unmarshal(payload, target)
}

func (f *fakeCacheStore) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if f.values == nil {
		f.values = map[string][]byte{}
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	f.values[key] = payload
	return nil
}

func (f *fakeCacheStore) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(f.values, key)
	}
	return nil
}

func (f *fakeCacheStore) DeleteByPrefix(ctx context.Context, prefix string) error {
	for key := range f.values {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(f.values, key)
		}
	}
	return nil
}

func (f *fakeSMSVerificationService) VerifyCode(ctx context.Context, input VerifySMSCodeInput) error {
	f.verified = append(f.verified, input)
	return f.verifyErr
}

func (f *fakeSMSVerificationService) SendCode(ctx context.Context, input SendSMSCodeInput) (SMSVerificationPayload, error) {
	f.sent = append(f.sent, input)
	return SMSVerificationPayload{
		Provider:   "mock",
		CooldownIn: 60,
		DebugCode:  "123456",
	}, f.sendErr
}

func (f *fakeEmailVerificationService) SendCode(ctx context.Context, input SendEmailCodeInput) (SMSVerificationPayload, error) {
	f.sent = append(f.sent, input)
	return SMSVerificationPayload{
		Provider:   "mock",
		CooldownIn: 60,
		DebugCode:  "654321",
	}, f.sendErr
}

func (f *fakeEmailVerificationService) VerifyCode(ctx context.Context, input VerifyEmailCodeInput) error {
	f.verified = append(f.verified, input)
	return f.verifyErr
}

func (f *fakeImageCaptchaService) Create(ctx context.Context) (ImageCaptchaPayload, error) {
	return ImageCaptchaPayload{
		CaptchaID: "captcha_123",
		ImageData: "data:image/svg+xml;base64,abc",
		ExpiresIn: 300,
	}, nil
}

func (f *fakeImageCaptchaService) Verify(ctx context.Context, captchaID string, code string) error {
	f.verified = append(f.verified, [2]string{captchaID, code})
	if captchaID == "" || code == "" {
		return utils.NewAppError(400, 400, "image captcha is required")
	}
	return f.verifyErr
}
