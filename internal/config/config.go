package config

import (
	"flag"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// Config хранит конфигурацию приложения.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

var (
	config     *Config
	configOnce sync.Once
)

// Load загружает конфигурацию однократно.
func Load() (*Config, error) {
	var loadErr error
	configOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			loadErr = err
			return
		}

		config = &Config{}
		flag.StringVar(&config.RunAddress, "a", os.Getenv("RUN_ADDRESS"), "server address")
		flag.StringVar(&config.DatabaseURI, "d", os.Getenv("DATABASE_URI"), "database URI")
		flag.StringVar(&config.AccrualSystemAddress, "r", os.Getenv("ACCRUAL_SYSTEM_ADDRESS"), "accrual system address")
		flag.Parse()

		if config.RunAddress == "" {
			config.RunAddress = "localhost:8080"
		}
		if config.DatabaseURI == "" {
			loadErr = os.ErrNotExist
			return
		}
		if config.AccrualSystemAddress == "" {
			config.AccrualSystemAddress = "http://localhost:8000"
		}
	})
	return config, loadErr
}
