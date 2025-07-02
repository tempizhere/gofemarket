package service

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/tempizhere/gofemarket/internal/api"
	"github.com/tempizhere/gofemarket/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// userServiceImpl реализует логику пользователей.
type userServiceImpl struct {
	repo      api.UserRepository
	jwtSecret string
}

// NewUserService создает новый UserService.
func NewUserService(repo api.UserRepository) api.UserService {
	return &userServiceImpl{
		repo:      repo,
		jwtSecret: "secret_key", // В реальном проекте использовать переменную окружения
	}
}

// Register регистрирует нового пользователя.
func (s *userServiceImpl) Register(ctx context.Context, login, password string) (string, error) {
	if login == "" || password == "" {
		return "", model.ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	userID, err := s.repo.CreateUserWithBalance(ctx, login, string(hashedPassword))
	if err != nil {
		return "", err
	}

	return s.generateToken(userID)
}

// Login аутентифицирует пользователя.
func (s *userServiceImpl) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", model.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", model.ErrInvalidCredentials
	}

	return s.generateToken(user.ID)
}

// ValidateToken проверяет JWT-токен.
func (s *userServiceImpl) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, model.ErrInvalidCredentials
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, model.ErrInvalidCredentials
		}
		return int(userID), nil
	}
	return 0, model.ErrInvalidCredentials
}

// generateToken создает JWT-токен.
func (s *userServiceImpl) generateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(s.jwtSecret))
}
