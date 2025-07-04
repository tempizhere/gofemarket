package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository/mocks"
)

// TestUserRepository_CreateUserWithBalance проверяет создание пользователя и баланса с различными сценариями (успех, дублирование логина, ошибка БД).
func TestUserRepository_CreateUserWithBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		login     string
		password  string
		setupMock func(*mocks.MockUserRepository)
		wantID    int
		wantErr   error
	}{
		{
			name:     "success",
			login:    "user1",
			password: "pass1",
			setupMock: func(mock *mocks.MockUserRepository) {
				mock.EXPECT().
					CreateUserWithBalance(gomock.Any(), "user1", "pass1").
					Return(1, nil)
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name:     "duplicate login",
			login:    "user2",
			password: "pass2",
			setupMock: func(mock *mocks.MockUserRepository) {
				mock.EXPECT().
					CreateUserWithBalance(gomock.Any(), "user2", "pass2").
					Return(0, model.ErrLoginTaken)
			},
			wantID:  0,
			wantErr: model.ErrLoginTaken,
		},
		{
			name:     "db error on balance insert",
			login:    "user3",
			password: "pass3",
			setupMock: func(mock *mocks.MockUserRepository) {
				mock.EXPECT().
					CreateUserWithBalance(gomock.Any(), "user3", "pass3").
					Return(0, errors.New("db error"))
			},
			wantID:  0,
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo)

			id, err := mockRepo.CreateUserWithBalance(ctx, tt.login, tt.password)
			assert.Equal(t, tt.wantID, id)
			if tt.wantErr != nil {
				if errors.Is(tt.wantErr, model.ErrLoginTaken) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserRepository_GetUserByLogin проверяет получение пользователя по логину с различными сценариями (успех, не найден, ошибка БД).
func TestUserRepository_GetUserByLogin(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		login     string
		setupMock func(*mocks.MockUserRepository)
		wantUser  model.User
		wantErr   error
	}{
		{
			name:  "success",
			login: "user1",
			setupMock: func(mock *mocks.MockUserRepository) {
				mock.EXPECT().
					GetUserByLogin(gomock.Any(), "user1").
					Return(model.User{ID: 1, Login: "user1", Password: "hashed_pass"}, nil)
			},
			wantUser: model.User{ID: 1, Login: "user1", Password: "hashed_pass"},
			wantErr:  nil,
		},
		{
			name:  "not found",
			login: "user2",
			setupMock: func(mock *mocks.MockUserRepository) {
				mock.EXPECT().
					GetUserByLogin(gomock.Any(), "user2").
					Return(model.User{}, model.ErrInvalidCredentials)
			},
			wantUser: model.User{},
			wantErr:  model.ErrInvalidCredentials,
		},
		{
			name:  "db error",
			login: "user3",
			setupMock: func(mock *mocks.MockUserRepository) {
				mock.EXPECT().
					GetUserByLogin(gomock.Any(), "user3").
					Return(model.User{}, errors.New("db error"))
			},
			wantUser: model.User{},
			wantErr:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockUserRepository(ctrl)
			tt.setupMock(mockRepo)

			user, err := mockRepo.GetUserByLogin(ctx, tt.login)
			assert.Equal(t, tt.wantUser, user)
			if tt.wantErr != nil {
				if errors.Is(tt.wantErr, model.ErrInvalidCredentials) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
