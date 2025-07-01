package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/tempizhere/gofemarket/internal/client"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/repository"
)

var stringPool = sync.Pool{
	New: func() interface{} {
		s := ""
		return &s
	},
}

// orderServiceImpl реализует логику заказов.
type orderServiceImpl struct {
	repo              *repository.OrderRepository
	accrualClient     *client.AccrualClient
	accrualSystemAddr *string
}

// NewOrderService создает новый OrderService.
func NewOrderService(repo *repository.OrderRepository, accrualSystemAddr string) OrderService {
	return &orderServiceImpl{
		repo:              repo,
		accrualClient:     client.NewAccrualClient(accrualSystemAddr),
		accrualSystemAddr: stringPool.Get().(*string),
	}
}

// UploadOrder загружает номер заказа.
func (s *orderServiceImpl) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	defer stringPool.Put(s.accrualSystemAddr)
	if !isValidLuhn(orderNumber) {
		return model.ErrInvalidOrderFormat
	}

	existingOrder, err := s.repo.GetOrderByNumber(ctx, orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			return model.ErrOrderAlreadyUploaded
		}
		return model.ErrOrderTakenByAnother
	}

	return s.repo.CreateOrder(ctx, userID, orderNumber, "NEW")
}

// GetOrders возвращает список заказов пользователя.
func (s *orderServiceImpl) GetOrders(ctx context.Context, userID int) ([]model.Order, error) {
	return s.repo.GetOrdersByUser(ctx, userID)
}

// ProcessOrders обрабатывает заказы в фоне.
func (s *orderServiceImpl) ProcessOrders(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			orders, err := s.repo.GetOrdersByStatus(ctx, []string{"NEW", "PROCESSING"})
			if err != nil {
				return err
			}
			for _, order := range orders {
				s.processOrder(ctx, &order)
			}
		}
	}
}

// processOrder обрабатывает один заказ.
func (s *orderServiceImpl) processOrder(ctx context.Context, order *model.Order) {
	accrual, err := s.accrualClient.GetAccrual(ctx, order.Number)
	if err != nil {
		if err == model.ErrOrderNotFound {
			return
		}
		if err == model.ErrTooManyRequests {
			time.Sleep(time.Second * 60)
			return
		}
		return
	}

	if accrual.Status == "INVALID" || accrual.Status == "PROCESSED" {
		_ = s.repo.UpdateOrderAndBalance(ctx, order.Number, accrual.Status, accrual.Accrual, order.UserID)
	}
}

// isValidLuhn проверяет номер заказа по алгоритму Луна.
func isValidLuhn(number string) bool {
	num, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return false
	}

	var sum int
	odd := true
	for num > 0 {
		digit := int(num % 10)
		if !odd {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		num /= 10
		odd = !odd
	}
	return sum%10 == 0
}
