package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/tempizhere/gofemarket/internal/model"
	"github.com/tempizhere/gofemarket/internal/service"
	"go.uber.org/zap"
)

// contextKey определяет тип для ключей контекста.
type contextKey string

// userIDKey ключ для userID в контексте.
const userIDKey contextKey = "userID"

// AuthMiddleware проверяет авторизацию пользователя.
func AuthMiddleware(userService service.UserService, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token != "" && strings.HasPrefix(token, "Bearer ") {
				token = strings.TrimPrefix(token, "Bearer ")
			} else if cookie, err := r.Cookie("token"); err == nil {
				token = cookie.Value
			}

			if token == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, model.ErrMsgInvalidCredentials, http.StatusUnauthorized)
				return
			}

			userID, err := userService.ValidateToken(token)
			if err != nil {
				logger.Error("invalid token", zap.Error(err))
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, model.ErrMsgInvalidCredentials, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
