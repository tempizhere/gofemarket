package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/service"
	"go.uber.org/zap"
)

// Handler структура для хранения зависимостей хендлеров.
type Handler struct {
	userService    service.UserService
	orderService   service.OrderService
	balanceService service.BalanceService
	logger         *zap.Logger
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(userService service.UserService, orderService service.OrderService, balanceService service.BalanceService, logger *zap.Logger) *Handler {
	return &Handler{
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
		logger:         logger,
	}
}

// Route структура для описания маршрута.
type Route struct {
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
	Protected   bool
}

// SetupRoutes настраивает маршруты API.
func SetupRoutes(router *mux.Router, userService service.UserService, orderService service.OrderService, balanceService service.BalanceService, logger *zap.Logger) {
	h := NewHandler(userService, orderService, balanceService, logger)
	routes := []Route{
		{Method: http.MethodPost, Path: "/api/user/register", HandlerFunc: h.register, Protected: false},
		{Method: http.MethodPost, Path: "/api/user/login", HandlerFunc: h.login, Protected: false},
		{Method: http.MethodPost, Path: "/api/user/orders", HandlerFunc: h.uploadOrder, Protected: true},
		{Method: http.MethodGet, Path: "/api/user/orders", HandlerFunc: h.getOrders, Protected: true},
		{Method: http.MethodGet, Path: "/api/user/balance", HandlerFunc: h.getBalance, Protected: true},
		{Method: http.MethodPost, Path: "/api/user/balance/withdraw", HandlerFunc: h.withdraw, Protected: true},
		{Method: http.MethodGet, Path: "/api/user/withdrawals", HandlerFunc: h.getWithdrawals, Protected: true},
	}

	for _, route := range routes {
		logger.Info("registering route", zap.String("method", route.Method), zap.String("path", route.Path))
		if route.Protected {
			router.Handle(route.Path, AuthMiddleware(userService, logger)(route.HandlerFunc)).Methods(route.Method)
		} else {
			router.HandleFunc(route.Path, route.HandlerFunc).Methods(route.Method)
		}
	}
}

// register обрабатывает регистрацию пользователя.
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, model.ErrMsgInvalidRequest, http.StatusBadRequest)
		return
	}

	token, err := h.userService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case model.ErrLoginTaken:
			http.Error(w, model.ErrMsgLoginTaken, http.StatusConflict)
		case model.ErrInvalidCredentials:
			http.Error(w, model.ErrMsgInvalidCredentials, http.StatusBadRequest)
		default:
			h.logger.Error("register error", zap.Error(err))
			http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})
	w.WriteHeader(http.StatusOK)
}

// login обрабатывает аутентификацию пользователя.
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, model.ErrMsgInvalidRequest, http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case model.ErrInvalidCredentials:
			http.Error(w, model.ErrMsgInvalidCredentials, http.StatusUnauthorized)
		default:
			h.logger.Error("login error", zap.Error(err))
			http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})
	w.WriteHeader(http.StatusOK)
}

// uploadOrder обрабатывает загрузку номера заказа.
func (h *Handler) uploadOrder(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received upload order request")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, model.ErrMsgInvalidRequest, http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	userID := r.Context().Value(userIDKey).(int)
	err = h.orderService.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch err {
		case model.ErrInvalidOrderFormat:
			http.Error(w, model.ErrMsgInvalidOrderFormat, http.StatusUnprocessableEntity)
		case model.ErrOrderAlreadyUploaded:
			http.Error(w, model.ErrMsgOrderAlreadyUploaded, http.StatusOK)
		case model.ErrOrderTakenByAnother:
			http.Error(w, model.ErrMsgOrderTakenByAnother, http.StatusConflict)
		default:
			h.logger.Error("upload order error", zap.Error(err))
			http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// getOrders возвращает список заказов пользователя.
func (h *Handler) getOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		h.logger.Error("get orders error", zap.Error(err))
		http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Error("encode orders error", zap.Error(err))
		http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
	}
}

// getBalance возвращает текущий баланс пользователя.
func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("get balance error", zap.Error(err))
		http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Error("encode balance error", zap.Error(err))
		http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
	}
}

// withdraw обрабатывает списание баллов.
func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, model.ErrMsgInvalidRequest, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(userIDKey).(int)
	err := h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case model.ErrInvalidOrderFormat:
			http.Error(w, model.ErrMsgInvalidOrderFormat, http.StatusUnprocessableEntity)
		case model.ErrInsufficientFunds:
			http.Error(w, model.ErrMsgInsufficientFunds, http.StatusPaymentRequired)
		default:
			h.logger.Error("withdraw error", zap.Error(err))
			http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

// getWithdrawals возвращает историю списаний.
func (h *Handler) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		h.logger.Error("get withdrawals error", zap.Error(err))
		http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Error("encode withdrawals error", zap.Error(err))
		http.Error(w, model.ErrMsgInternalServer, http.StatusInternalServerError)
	}
}
