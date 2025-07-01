package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/tempizhere/gofemarket/internal/api"
	"github.com/tempizhere/gofemarket/internal/config"
	"github.com/tempizhere/gofemarket/internal/logger"
	"github.com/tempizhere/gofemarket/internal/repository"
	"github.com/tempizhere/gofemarket/internal/service"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// main инициализирует и запускает сервер.
func main() {
	// Инициализация логгера
	log, err := logger.NewLogger()
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer func() {
		if err := log.Sync(); err != nil {
			log.Error("failed to sync logger", zap.Error(err))
		}
	}()

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	// Инициализация подключения к БД
	db, err := repository.NewDB(cfg.DatabaseURI)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("failed to close database", zap.Error(err))
		}
	}()

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	orderService := service.NewOrderService(orderRepo, cfg.AccrualSystemAddress)
	balanceService := service.NewBalanceService(balanceRepo, orderRepo)

	// Настройка маршрутов
	router := mux.NewRouter()
	api.SetupRoutes(router, userService, orderService, balanceService, log)

	// Создание HTTP-сервера
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	// Контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Запуск сервера и фоновой обработки заказов
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		log.Info("starting server", zap.String("address", cfg.RunAddress))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		return orderService.ProcessOrders(gCtx)
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Info("shutting down server")
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}
