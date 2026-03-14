package service

import (
	"context"
	"errors"
	"time"

	"github.com/gee-coder/template-go-backend/internal/repository"
	"github.com/gee-coder/template-go-backend/internal/repository/model"
	"github.com/gee-coder/template-go-backend/internal/utils"
)

type fakeUserRepository struct {
	users        map[string]*model.User
	usersByID    map[uint]*model.User
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
	user.ID = uint(len(f.users) + 1)
	f.users[user.Username] = user
	return nil
}

func (f *fakeUserRepository) Update(ctx context.Context, user *model.User) error {
	if _, ok := f.users[user.Username]; !ok {
		return utils.ErrNotFound
	}
	f.users[user.Username] = user
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

