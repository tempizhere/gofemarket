package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository/mocks"
)

// TestOrderService_UploadOrder проверяет загрузку заказа пользователем через сервис.
func TestOrderService_UploadOrder(t *testing.T) {
	tests := []struct {
		name        string
		orderNumber string
		userID      int
		repoOrder   model.Order
		repoErr     error
		wantErr     error
	}{
		{
			name:        "valid order",
			orderNumber: "12345678903",
			userID:      1,
			repoErr:     sql.ErrNoRows,
			wantErr:     nil,
		},
		{
			name:        "invalid order format",
			orderNumber: "123",
			userID:      1,
			wantErr:     model.ErrInvalidOrderFormat,
		},
		{
			name:        "order already uploaded",
			orderNumber: "12345678903",
			userID:      1,
			repoOrder:   model.Order{UserID: 1},
			repoErr:     nil,
			wantErr:     model.ErrOrderAlreadyUploaded,
		},
		{
			name:        "order taken by another",
			orderNumber: "12345678903",
			userID:      1,
			repoOrder:   model.Order{UserID: 2},
			repoErr:     nil,
			wantErr:     model.ErrOrderTakenByAnother,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			repo := mocks.NewMockOrderRepository(ctrl)
			s := NewOrderService(repo, "http://localhost").(*orderServiceImpl)

			if tt.wantErr != model.ErrInvalidOrderFormat {
				repo.EXPECT().GetOrderByNumber(ctx, tt.orderNumber).Return(tt.repoOrder, tt.repoErr).Times(1)
			}
			if tt.wantErr == nil {
				repo.EXPECT().CreateOrder(ctx, tt.userID, tt.orderNumber, "NEW").Return(nil).Times(1)
			}
			err := s.UploadOrder(ctx, tt.userID, tt.orderNumber)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

// TestOrderService_GetOrders проверяет получение всех заказов пользователя через сервис.
func TestOrderService_GetOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	repo := mocks.NewMockOrderRepository(ctrl)
	s := NewOrderService(repo, "http://localhost").(*orderServiceImpl)

	expectedOrders := []model.Order{{Number: "12345678903", UserID: 1}}
	repo.EXPECT().GetOrdersByUser(ctx, 1).Return(expectedOrders, nil).Times(1)

	orders, err := s.GetOrders(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)

	// Тест для случая с ошибкой
	t.Run("error case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		repo := mocks.NewMockOrderRepository(ctrl)
		s := NewOrderService(repo, "http://localhost").(*orderServiceImpl)

		repo.EXPECT().GetOrdersByUser(ctx, 1).Return(nil, errors.New("db error")).Times(1)
		_, err := s.GetOrders(ctx, 1)
		assert.Error(t, err)
	})
}

// TestOrderService_ProcessOrder проверяет обработку одного заказа через сервис.
func TestOrderService_ProcessOrder(t *testing.T) {
	tests := []struct {
		name      string
		order     model.Order
		clientErr error
		updateErr error
		wantErr   error
	}{
		{
			name:      "success processed",
			order:     model.Order{Number: "12345678903", UserID: 1, Status: "NEW"},
			clientErr: nil,
			updateErr: nil,
			wantErr:   nil,
		},
		{
			name:      "success invalid",
			order:     model.Order{Number: "12345678903", UserID: 1, Status: "NEW"},
			clientErr: nil,
			updateErr: nil,
			wantErr:   nil,
		},
		{
			name:      "order not found",
			order:     model.Order{Number: "12345678903", UserID: 1, Status: "NEW"},
			clientErr: model.ErrOrderNotFound,
			updateErr: nil,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := context.Background()
			repo := mocks.NewMockOrderRepository(ctrl)
			client := mocks.NewMockAccrualClient(ctrl)
			s := &orderServiceImpl{
				repo:              repo,
				accrualClient:     client,
				accrualSystemAddr: "http://localhost",
			}

			// Настройка моков
			accrual := model.Order{Number: "12345678903", Status: "PROCESSED", Accrual: 10.0}
			if tt.name == "success invalid" {
				accrual.Status = "INVALID"
			}
			client.EXPECT().GetAccrual(gomock.Any(), tt.order.Number).Return(accrual, time.Duration(0), tt.clientErr).Times(1)
			if tt.clientErr == nil && (accrual.Status == "PROCESSED" || accrual.Status == "INVALID") {
				repo.EXPECT().UpdateOrderAndBalance(gomock.Any(), tt.order.Number, accrual.Status, accrual.Accrual, tt.order.UserID).Return(tt.updateErr).Times(1)
			}

			err := s.processOrder(ctx, &tt.order)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
