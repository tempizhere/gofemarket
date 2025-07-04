package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository/mocks"
)

// TestOrderRepository_CreateOrder проверяет создание заказа с различными сценариями (успех, дублирование номера).
func TestOrderRepository_CreateOrder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		userID    int
		number    string
		status    string
		setupMock func(*mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			name:   "success",
			userID: 1,
			number: "123",
			status: "NEW",
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					CreateOrder(gomock.Any(), 1, "123", "NEW").
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "duplicate number",
			userID: 2,
			number: "456",
			status: "NEW",
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					CreateOrder(gomock.Any(), 2, "456", "NEW").
					Return(&pq.Error{Code: "23505"})
			},
			wantErr: &pq.Error{Code: "23505"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.CreateOrder(ctx, tt.userID, tt.number, tt.status)
			if tt.wantErr != nil {
				var pqErr *pq.Error
				if errors.As(tt.wantErr, &pqErr) {
					var actualPQErr *pq.Error
					require.ErrorAs(t, err, &actualPQErr)
					assert.Equal(t, pqErr.Code, actualPQErr.Code)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrderRepository_GetOrderByNumber проверяет получение заказа по номеру (успех, не найден, ошибка БД).
func TestOrderRepository_GetOrderByNumber(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		number    string
		setupMock func(*mocks.MockOrderRepository)
		wantOrder model.Order
		wantErr   error
	}{
		{
			name:   "success",
			number: "123",
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrderByNumber(gomock.Any(), "123").
					Return(model.Order{
						ID:         1,
						Number:     "123",
						UserID:     1,
						Status:     "NEW",
						Accrual:    100.0,
						UploadedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
			},
			wantOrder: model.Order{
				ID:         1,
				Number:     "123",
				UserID:     1,
				Status:     "NEW",
				Accrual:    100.0,
				UploadedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: nil,
		},
		{
			name:   "not found",
			number: "456",
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrderByNumber(gomock.Any(), "456").
					Return(model.Order{}, sql.ErrNoRows)
			},
			wantOrder: model.Order{},
			wantErr:   sql.ErrNoRows,
		},
		{
			name:   "db error",
			number: "789",
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrderByNumber(gomock.Any(), "789").
					Return(model.Order{}, errors.New("db error"))
			},
			wantOrder: model.Order{},
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			tt.setupMock(mockRepo)

			order, err := mockRepo.GetOrderByNumber(ctx, tt.number)
			assert.Equal(t, tt.wantOrder, order)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrderRepository_GetOrdersByUser проверяет получение заказов пользователя (успех, нет заказов, ошибка БД).
func TestOrderRepository_GetOrdersByUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		userID     int
		setupMock  func(*mocks.MockOrderRepository)
		wantOrders []model.Order
		wantErr    error
	}{
		{
			name:   "success with data",
			userID: 1,
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrdersByUser(gomock.Any(), 1).
					Return([]model.Order{
						{ID: 1, Number: "123", Status: "NEW", Accrual: 100.0, UploadedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
						{ID: 2, Number: "456", Status: "PROCESSED", Accrual: 200.0, UploadedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
					}, nil)
			},
			wantOrders: []model.Order{
				{ID: 1, Number: "123", Status: "NEW", Accrual: 100.0, UploadedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{ID: 2, Number: "456", Status: "PROCESSED", Accrual: 200.0, UploadedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
			wantErr: nil,
		},
		{
			name:   "no orders",
			userID: 2,
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrdersByUser(gomock.Any(), 2).
					Return([]model.Order{}, nil)
			},
			wantOrders: []model.Order{},
			wantErr:    nil,
		},
		{
			name:   "db error",
			userID: 3,
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrdersByUser(gomock.Any(), 3).
					Return(nil, errors.New("db error"))
			},
			wantOrders: nil,
			wantErr:    errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			tt.setupMock(mockRepo)

			orders, err := mockRepo.GetOrdersByUser(ctx, tt.userID)
			assert.Equal(t, tt.wantOrders, orders)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrderRepository_GetOrdersByStatus проверяет получение заказов по статусам (успех, нет заказов, ошибка БД).
func TestOrderRepository_GetOrdersByStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		statuses   []string
		setupMock  func(*mocks.MockOrderRepository)
		wantOrders []model.Order
		wantErr    error
	}{
		{
			name:     "success with data",
			statuses: []string{"NEW", "PROCESSING"},
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrdersByStatus(gomock.Any(), []string{"NEW", "PROCESSING"}).
					Return([]model.Order{
						{ID: 1, Number: "123", UserID: 1, Status: "NEW", Accrual: 100.0, UploadedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
						{ID: 2, Number: "456", UserID: 2, Status: "PROCESSING", Accrual: 200.0, UploadedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
					}, nil)
			},
			wantOrders: []model.Order{
				{ID: 1, Number: "123", UserID: 1, Status: "NEW", Accrual: 100.0, UploadedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{ID: 2, Number: "456", UserID: 2, Status: "PROCESSING", Accrual: 200.0, UploadedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
			},
			wantErr: nil,
		},
		{
			name:     "no orders",
			statuses: []string{"INVALID"},
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrdersByStatus(gomock.Any(), []string{"INVALID"}).
					Return([]model.Order{}, nil)
			},
			wantOrders: []model.Order{},
			wantErr:    nil,
		},
		{
			name:     "db error",
			statuses: []string{"NEW"},
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					GetOrdersByStatus(gomock.Any(), []string{"NEW"}).
					Return(nil, errors.New("db error"))
			},
			wantOrders: nil,
			wantErr:    errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			tt.setupMock(mockRepo)

			orders, err := mockRepo.GetOrdersByStatus(ctx, tt.statuses)
			assert.Equal(t, tt.wantOrders, orders)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrderRepository_UpdateOrderAndBalance проверяет обновление заказа и баланса (успех с начислением, успех без начисления, ошибка БД).
func TestOrderRepository_UpdateOrderAndBalance(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		number    string
		status    string
		accrual   float64
		userID    int
		setupMock func(*mocks.MockOrderRepository)
		wantErr   error
	}{
		{
			name:    "success with accrual",
			number:  "123",
			status:  "PROCESSED",
			accrual: 100.0,
			userID:  1,
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					UpdateOrderAndBalance(gomock.Any(), "123", "PROCESSED", 100.0, 1).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "success without accrual",
			number:  "456",
			status:  "INVALID",
			accrual: 0.0,
			userID:  2,
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					UpdateOrderAndBalance(gomock.Any(), "456", "INVALID", 0.0, 2).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "db error on order update",
			number:  "789",
			status:  "PROCESSED",
			accrual: 100.0,
			userID:  3,
			setupMock: func(mock *mocks.MockOrderRepository) {
				mock.EXPECT().
					UpdateOrderAndBalance(gomock.Any(), "789", "PROCESSED", 100.0, 3).
					Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockOrderRepository(ctrl)
			tt.setupMock(mockRepo)

			err := mockRepo.UpdateOrderAndBalance(ctx, tt.number, tt.status, tt.accrual, tt.userID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
