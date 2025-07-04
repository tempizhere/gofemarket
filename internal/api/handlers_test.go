package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/tempizhere/gofemarket/internal/model"
	"go.uber.org/zap"
)

// mockUserService мокает UserService.
type mockUserService struct {
	registerFn func(ctx context.Context, login, password string) (string, error)
	loginFn    func(ctx context.Context, login, password string) (string, error)
	validateFn func(token string) (int, error)
}

func (m *mockUserService) Register(ctx context.Context, login, password string) (string, error) {
	return m.registerFn(ctx, login, password)
}

func (m *mockUserService) Login(ctx context.Context, login, password string) (string, error) {
	return m.loginFn(ctx, login, password)
}

func (m *mockUserService) ValidateToken(token string) (int, error) {
	return m.validateFn(token)
}

// mockOrderService мокает OrderService.
type mockOrderService struct {
	uploadOrderFn func(ctx context.Context, userID int, orderNumber string) error
	getOrdersFn   func(ctx context.Context, userID int) ([]model.Order, error)
}

func (m *mockOrderService) UploadOrder(ctx context.Context, userID int, orderNumber string) error {
	return m.uploadOrderFn(ctx, userID, orderNumber)
}

func (m *mockOrderService) GetOrders(ctx context.Context, userID int) ([]model.Order, error) {
	return m.getOrdersFn(ctx, userID)
}

func (m *mockOrderService) ProcessOrders(ctx context.Context) error {
	return nil
}

// mockBalanceService мокает BalanceService.
type mockBalanceService struct {
	getBalanceFn     func(ctx context.Context, userID int) (model.Balance, error)
	withdrawFn       func(ctx context.Context, userID int, orderNumber string, sum float64) error
	getWithdrawalsFn func(ctx context.Context, userID int) ([]model.Withdrawal, error)
}

func (m *mockBalanceService) GetBalance(ctx context.Context, userID int) (model.Balance, error) {
	return m.getBalanceFn(ctx, userID)
}

func (m *mockBalanceService) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	return m.withdrawFn(ctx, userID, orderNumber, sum)
}

func (m *mockBalanceService) GetWithdrawals(ctx context.Context, userID int) ([]model.Withdrawal, error) {
	return m.getWithdrawalsFn(ctx, userID)
}

// TestHandlers тестирует все хендлеры API.
func TestHandlers(t *testing.T) {
	logger, _ := zap.NewProduction() // Минимальный логгер для тестов
	userService := &mockUserService{}
	orderService := &mockOrderService{}
	balanceService := &mockBalanceService{}
	router := mux.NewRouter()
	SetupRoutes(router, userService, orderService, balanceService, logger)

	t.Run("POST /api/user/register success", func(t *testing.T) {
		userService.registerFn = func(ctx context.Context, login, password string) (string, error) {
			return "token", nil
		}
		body := `{"login":"testuser","password":"testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if cookie := w.Header().Get("Set-Cookie"); !strings.Contains(cookie, "token=token") {
			t.Errorf("expected cookie with token, got %s", cookie)
		}
	})

	t.Run("POST /api/user/register login taken", func(t *testing.T) {
		userService.registerFn = func(ctx context.Context, login, password string) (string, error) {
			return "", model.ErrLoginTaken
		}
		body := `{"login":"testuser","password":"testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
		}
	})

	t.Run("POST /api/user/register invalid format", func(t *testing.T) {
		body := `{"login":"testuser","password":}` // Неверный JSON
		req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("POST /api/user/login success", func(t *testing.T) {
		userService.loginFn = func(ctx context.Context, login, password string) (string, error) {
			return "token", nil
		}
		body := `{"login":"testuser","password":"testpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if cookie := w.Header().Get("Set-Cookie"); !strings.Contains(cookie, "token=token") {
			t.Errorf("expected cookie with token, got %s", cookie)
		}
	})

	t.Run("POST /api/user/login invalid credentials", func(t *testing.T) {
		userService.loginFn = func(ctx context.Context, login, password string) (string, error) {
			return "", model.ErrInvalidCredentials
		}
		body := `{"login":"testuser","password":"wrongpass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("POST /api/user/login invalid format", func(t *testing.T) {
		body := `{"login":"testuser","password":}` // Неверный JSON
		req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("POST /api/user/orders success", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		orderService.uploadOrderFn = func(ctx context.Context, userID int, orderNumber string) error {
			return nil
		}
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
		req.Header.Set("Content-Type", "text/plain")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
		}
	})

	t.Run("POST /api/user/orders already uploaded", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		orderService.uploadOrderFn = func(ctx context.Context, userID int, orderNumber string) error {
			return model.ErrOrderAlreadyUploaded
		}
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
		req.Header.Set("Content-Type", "text/plain")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("POST /api/user/orders unauthorized", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 0, model.ErrInvalidCredentials
		}
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("12345678903"))
		req.Header.Set("Content-Type", "text/plain")
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("POST /api/user/orders invalid format", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		orderService.uploadOrderFn = func(ctx context.Context, userID int, orderNumber string) error {
			return model.ErrInvalidOrderFormat
		}
		req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString("123"))
		req.Header.Set("Content-Type", "text/plain")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("GET /api/user/orders success", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		orderService.getOrdersFn = func(ctx context.Context, userID int) ([]model.Order, error) {
			return []model.Order{
				{Number: "12345678903", Status: "NEW", UploadedAt: time.Now()},
			}, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		var orders []model.Order
		if err := json.NewDecoder(w.Body).Decode(&orders); err != nil {
			t.Errorf("failed to decode response: %v", err)
		}
		if len(orders) != 1 || orders[0].Number != "12345678903" {
			t.Errorf("unexpected orders: %+v", orders)
		}
	})

	t.Run("GET /api/user/orders no content", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		orderService.getOrdersFn = func(ctx context.Context, userID int) ([]model.Order, error) {
			return []model.Order{}, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})

	t.Run("GET /api/user/orders unauthorized", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 0, model.ErrInvalidCredentials
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("GET /api/user/balance success", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		balanceService.getBalanceFn = func(ctx context.Context, userID int) (model.Balance, error) {
			return model.Balance{Current: 500, Withdrawn: 100}, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		var balance model.Balance
		if err := json.NewDecoder(w.Body).Decode(&balance); err != nil {
			t.Errorf("failed to decode response: %v", err)
		}
		if balance.Current != 500 || balance.Withdrawn != 100 {
			t.Errorf("unexpected balance: %+v", balance)
		}
	})

	t.Run("GET /api/user/balance unauthorized", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 0, model.ErrInvalidCredentials
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("POST /api/user/balance/withdraw success", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		balanceService.withdrawFn = func(ctx context.Context, userID int, orderNumber string, sum float64) error {
			return nil
		}
		body := `{"order":"9278923470","sum":100}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("POST /api/user/balance/withdraw insufficient funds", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		balanceService.withdrawFn = func(ctx context.Context, userID int, orderNumber string, sum float64) error {
			return model.ErrInsufficientFunds
		}
		body := `{"order":"9278923470","sum":100}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusPaymentRequired {
			t.Errorf("expected status %d, got %d", http.StatusPaymentRequired, w.Code)
		}
	})

	t.Run("POST /api/user/balance/withdraw invalid format", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		balanceService.withdrawFn = func(ctx context.Context, userID int, orderNumber string, sum float64) error {
			return model.ErrInvalidOrderFormat
		}
		body := `{"order":"123","sum":100}`
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}
	})

	t.Run("POST /api/user/balance/withdraw invalid request", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		body := `{"order":"9278923470","sum":}` // Неверный JSON
		req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("GET /api/user/withdrawals success", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		balanceService.getWithdrawalsFn = func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
			return []model.Withdrawal{
				{Order: "9278923470", Sum: 100, ProcessedAt: time.Now()},
			}, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		var withdrawals []model.Withdrawal
		if err := json.NewDecoder(w.Body).Decode(&withdrawals); err != nil {
			t.Errorf("failed to decode response: %v", err)
		}
		if len(withdrawals) != 1 || withdrawals[0].Order != "9278923470" {
			t.Errorf("unexpected withdrawals: %+v", withdrawals)
		}
	})

	t.Run("GET /api/user/withdrawals no content", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 1, nil
		}
		balanceService.getWithdrawalsFn = func(ctx context.Context, userID int) ([]model.Withdrawal, error) {
			return []model.Withdrawal{}, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "valid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}
	})

	t.Run("GET /api/user/withdrawals unauthorized", func(t *testing.T) {
		userService.validateFn = func(token string) (int, error) {
			return 0, model.ErrInvalidCredentials
		}
		req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}
