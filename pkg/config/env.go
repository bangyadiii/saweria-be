package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	JWTSecret            string
	JWTRefreshSecret     string
	JWTExpiryHours       int
	JWTRefreshExpiryDays int
	GoogleClientID       string
	GoogleClientSecret   string
	MidtransServerKey    string
	MidtransClientKey    string
	MidtransEnvironment  string
	PlatformFeePercent   float64
	BaseURL              string
	FrontendURL          string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		return nil, fmt.Errorf("JWT_REFRESH_SECRET is required")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	midtransServerKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if midtransServerKey == "" {
		return nil, fmt.Errorf("MIDTRANS_SERVER_KEY is required")
	}
	midtransClientKey := os.Getenv("MIDTRANS_CLIENT_KEY")
	if midtransClientKey == "" {
		return nil, fmt.Errorf("MIDTRANS_CLIENT_KEY is required")
	}

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          dbURL,
		JWTSecret:            jwtSecret,
		JWTRefreshSecret:     jwtRefreshSecret,
		JWTExpiryHours:       getEnvInt("JWT_EXPIRY_HOURS", 1),
		JWTRefreshExpiryDays: getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7),
		GoogleClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		MidtransServerKey:    midtransServerKey,
		MidtransClientKey:    midtransClientKey,
		MidtransEnvironment:  getEnv("MIDTRANS_ENVIRONMENT", "sandbox"),
		PlatformFeePercent:   getEnvFloat("PLATFORM_FEE_PERCENT", 5),
		BaseURL:              getEnv("BASE_URL", "http://localhost:8080"),
		FrontendURL:          getEnv("FRONTEND_URL", "http://localhost:3000"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
