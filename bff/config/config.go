package config

import (
	"os"
	"time"
)

// Config holds the BFF gateway configuration.
type Config struct {
	Server   Server
	GRPC     GRPC
	JWT      JWT
}

// Server config.
type Server struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// GRPC targets.
type GRPC struct {
	OrderServiceAddr   string
	ProductServiceAddr string
	CartServiceAddr    string
	UserServiceAddr    string
	PaymentServiceAddr string
	DialTimeout        time.Duration
}

// JWT config.
type JWT struct {
	Secret     string
	Expiration time.Duration
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		Server: Server{
			Addr:         getEnv("BFF_ADDR", ":8080"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		GRPC: GRPC{
			OrderServiceAddr:   getEnv("ORDER_SERVICE_ADDR", "127.0.0.1:9000"),
			ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "127.0.0.1:9001"),
			CartServiceAddr:    getEnv("CART_SERVICE_ADDR", "127.0.0.1:9002"),
			UserServiceAddr:    getEnv("USER_SERVICE_ADDR", "127.0.0.1:9003"),
			PaymentServiceAddr: getEnv("PAYMENT_SERVICE_ADDR", "127.0.0.1:9004"),
			DialTimeout:        5 * time.Second,
		},
		JWT: JWT{
			Secret:     getEnv("JWT_SECRET", "change-me-in-production"),
			Expiration: 24 * time.Hour,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
