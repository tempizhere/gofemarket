package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tempizhere/gofemarket/internal/model"
	"go.uber.org/zap"
)

// Handler структура для хранения зависимостей хендлеров.
type Handler struct {
	userService    UserService
	orderService   OrderService
	balanceService BalanceService
	logger         *zap.Logger
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(userService UserService, orderService OrderService, balanceService BalanceService, logger *zap.Logger) *Handler {
	return &Handler{
		userService:    userService,
		orderService:   orderService,
		balanceService: balanceService,
		logger:         logger,
	}
}

// SetupRoutes настраивает маршруты API.
func SetupRoutes(router *mux.Router, userService UserService, orderService OrderService, balanceService BalanceService, logger *zap.Logger) {
	h := NewHandler(userService, orderService, balanceService, logger)

	// Открытые маршруты
	router.HandleFunc("/api/user/register", h.register).Methods(http.MethodPost)
	router.HandleFunc("/api/user/login", h.login).Methods(http.MethodPost)

	// Защищённые маршруты через Subrouter
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(AuthMiddleware(userService, logger))
	protected.HandleFunc("/user/orders", h.uploadOrder).Methods(http.MethodPost)
	protected.HandleFunc("/user/orders", h.getOrders).Methods(http.MethodGet)
	protected.HandleFunc("/user/balance", h.getBalance).Methods(http.MethodGet)
	protected.HandleFunc("/user/balance/withdraw", h.withdraw).Methods(http.MethodPost)
	protected.HandleFunc("/user/withdrawals", h.getWithdrawals).Methods(http.MethodGet)

	logger.Info("registered routes")
}

// register обрабатывает регистрацию пользователя.
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := h.userService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case model.ErrLoginTaken:
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		case model.ErrInvalidCredentials:
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		default:
			h.logger.Error("register error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case model.ErrInvalidCredentials:
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		default:
			h.logger.Error("login error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	userID := r.Context().Value(userIDKey).(int)
	err = h.orderService.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch err {
		case model.ErrInvalidOrderFormat:
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case model.ErrOrderAlreadyUploaded:
			http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
		case model.ErrOrderTakenByAnother:
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		default:
			h.logger.Error("upload order error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Error("encode orders error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// getBalance возвращает текущий баланс пользователя.
func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("get balance error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Error("encode balance error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// withdraw обрабатывает списание баллов.
func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(userIDKey).(int)
	err := h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch err {
		case model.ErrInvalidOrderFormat:
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		case model.ErrInsufficientFunds:
			http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
		default:
			h.logger.Error("withdraw error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Error("encode withdrawals error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
