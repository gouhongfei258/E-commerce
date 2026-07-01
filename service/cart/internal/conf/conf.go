package conf

import (
	"os"
	"time"
)

type Config struct {
	Server   Server
	Database Database
}

type Server struct {
	Name    string
	Version string
	GRPC    GRPC
}

type GRPC struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Database struct {
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Server: Server{
			Name:    getEnv("SERVICE_NAME", "cart-service"),
			Version: getEnv("SERVICE_VERSION", "v1.0.0"),
			GRPC: GRPC{
				Addr:         getEnv("CART_GRPC_ADDR", ":9002"),
				ReadTimeout:  10 * time.Second,
				WriteTimeout: 10 * time.Second,
			},
		},
		Database: Database{
			DSN:             getEnv("DB_DSN", "root:@tcp(127.0.0.1:3306)/mall?charset=utf8mb4&parseTime=True&loc=Local"),
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: 30 * time.Minute,
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
