package main

import (
	"context"
	"log"
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
	logg, err := logger.NewLogger()
	if err != nil {
		log.Fatal("failed to init logger: " + err.Error())
	}
	defer func() {
		if err := logg.Sync(); err != nil {
			logg.Error("failed to sync logger", zap.Error(err))
		}
	}()

	logg.Info("server binary path", zap.String("path", os.Args[0]))

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logg.Fatal("failed to load config", zap.Error(err))
	}

	// Инициализация подключения к БД
	db, err := repository.NewDB(cfg.DatabaseURI)
	if err != nil {
		logg.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logg.Error("failed to close database", zap.Error(err))
		}
	}()

	// Инициализация таблиц
	if err := repository.InitTables(db); err != nil {
		logg.Fatal("failed to initialize tables", zap.Error(err))
	}

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
	api.SetupRoutes(router, userService, orderService, balanceService, logg)

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
		logg.Info("starting server", zap.String("address", cfg.RunAddress))
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
		logg.Info("shutting down server")
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		logg.Fatal("server error", zap.Error(err))
	}
}
