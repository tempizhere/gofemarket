package service

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository/mocks"
	"golang.org/x/crypto/bcrypt"
)

// TestUserService_Register проверяет регистрацию пользователя через сервис.
func TestUserService_Register(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		repoID   int
		repoErr  error
		wantErr  error
	}{
		{
			name:     "valid registration",
			login:    "user",
			password: "pass",
			repoID:   1,
			repoErr:  nil,
			wantErr:  nil,
		},
		{
			name:     "empty credentials",
			login:    "",
			password: "",
			wantErr:  model.ErrInvalidCredentials,
		},
		{
			name:     "repo error",
			login:    "user",
			password: "pass",
			repoErr:  model.ErrLoginTaken,
			wantErr:  model.ErrLoginTaken,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			repo := mocks.NewMockUserRepository(ctrl)
			s := NewUserService(repo).(*userServiceImpl)

			if tt.wantErr != model.ErrInvalidCredentials {
				repo.EXPECT().CreateUserWithBalance(ctx, tt.login, gomock.Any()).Return(tt.repoID, tt.repoErr).Times(1)
			}

			token, err := s.Register(ctx, tt.login, tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "error mismatch for test %s", tt.name)
				assert.Empty(t, token, "token should be empty for test %s", tt.name)
			} else {
				assert.NoError(t, err, "unexpected error for test %s", tt.name)
				assert.NotEmpty(t, token, "token should not be empty for test %s", tt.name)
			}
		})
	}
}

// TestUserService_Login проверяет вход пользователя (логин) через сервис.
func TestUserService_Login(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		user     model.User
		repoErr  error
		wantErr  error
	}{
		{
			name:     "valid login",
			login:    "user",
			password: "pass",
			user: model.User{ID: 1, Login: "user", Password: func() string {
				h, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
				return string(h)
			}()},
			repoErr: nil,
			wantErr: nil,
		},
		{
			name:     "invalid login",
			login:    "user",
			password: "pass",
			repoErr:  model.ErrInvalidCredentials,
			wantErr:  model.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			login:    "user",
			password: "wrong",
			user: model.User{ID: 1, Login: "user", Password: func() string {
				h, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
				return string(h)
			}()},
			repoErr: nil,
			wantErr: model.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			repo := mocks.NewMockUserRepository(ctrl)
			s := NewUserService(repo).(*userServiceImpl)

			repo.EXPECT().GetUserByLogin(ctx, tt.login).Return(tt.user, tt.repoErr).Times(1)
			token, err := s.Login(ctx, tt.login, tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "error mismatch for test %s", tt.name)
				assert.Empty(t, token, "token should be empty for test %s", tt.name)
			} else {
				assert.NoError(t, err, "unexpected error for test %s", tt.name)
				assert.NotEmpty(t, token, "token should not be empty for test %s", tt.name)
			}
		})
	}
}

// TestUserService_ValidateToken проверяет валидацию JWT токена пользователя через сервис.
func TestUserService_ValidateToken(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepository(ctrl)
	s := NewUserService(repo).(*userServiceImpl)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("secret_key"))

	tests := []struct {
		name       string
		token      string
		wantUserID int
		wantErr    bool
	}{
		{
			name:       "valid token",
			token:      tokenString,
			wantUserID: 1,
			wantErr:    false,
		},
		{
			name:       "invalid token",
			token:      "invalid",
			wantUserID: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			userID, err := s.ValidateToken(tt.token)
			if tt.wantErr {
				assert.Error(t, err, "expected error for test %s", tt.name)
			} else {
				assert.NoError(t, err, "unexpected error for test %s", tt.name)
			}
			assert.Equal(t, tt.wantUserID, userID, "userID mismatch for test %s", tt.name)
		})
	}
}
