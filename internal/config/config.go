package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

var C Config

type Config struct {
	Server         ServerConfig
	DB             DBConfig
	Auth           AuthConfig
	CORS           CORSConfig
	Rate           RateConfig
	Email          EmailConfig
	Paths          PathsConfig
	Zeitnahme      ZeitnahmeConfig
	GotenbergURL   string
	AppInternalURL string
	AppPublicURL   string
	Env            string
	Log            LogConfig
}

type LogConfig struct {
	Level string
}

type ZeitnahmeConfig struct {
	CurrentTag string
}

func (z ZeitnahmeConfig) GetCurrentTag() string {
	if z.CurrentTag != "" {
		return z.CurrentTag
	}
	switch time.Now().Weekday() {
	case time.Friday, time.Saturday:
		return "sa"
	case time.Sunday, time.Monday:
		return "so"
	default:
		return "sa"
	}
}

type ServerConfig struct {
	Port           string
	Host           string
	WSPort         string
	TrustedProxies []string
}

type DBConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	Name               string
	SSLMode            string
	PoolMaxConns       int32
	PoolMinConns       int32
	PoolMaxIdleSeconds int
	ConnectTimeoutSec  int
}

type AuthConfig struct {
	JWTSecret string
}

type CORSConfig struct {
	AllowedOrigins string
	AllowedMethods string
	AllowedHeaders string
}

type RateConfig struct {
	RPS            float64
	Burst          int
	UserMultiplier int
}

type EmailConfig struct {
	Sender   string
	Password string
	SMTPHost string
	SMTPPort int
}

type PathsConfig struct {
	UploadDir string
	FilesDir  string
	PublicDir string
}

func Load() {
	C = Config{
		Server: ServerConfig{
			Port:           getEnv("PORT", "3000"),
			Host:           getEnv("HOST", "127.0.0.1"),
			WSPort:         getEnv("WS_PORT", "8081"),
			TrustedProxies: getEnvList("TRUSTED_PROXIES"),
		},
		DB: DBConfig{
			Host:               os.Getenv("DB_HOST"),
			Port:               os.Getenv("DB_PORT"),
			User:               os.Getenv("DB_USER"),
			Password:           os.Getenv("DB_PASSWORD"),
			Name:               os.Getenv("DB_NAME"),
			SSLMode:            os.Getenv("DB_SSLMODE"),
			PoolMaxConns:       int32(getEnvInt("DB_POOL_MAX_CONNS", 20)),
			PoolMinConns:       int32(getEnvInt("DB_POOL_MIN_CONNS", 1)),
			PoolMaxIdleSeconds: getEnvInt("DB_POOL_MAX_IDLE_SECONDS", 300),
			ConnectTimeoutSec:  getEnvInt("DB_CONNECT_TIMEOUT_SECONDS", 10),
		},
		Auth: AuthConfig{
			JWTSecret: getEnvRequired("JWT_SECRET"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
			AllowedMethods: getEnv("CORS_ALLOWED_METHODS", "GET, POST, PUT, DELETE, OPTIONS"),
			AllowedHeaders: getEnv("CORS_ALLOWED_HEADERS", "Content-Type, Authorization"),
		},
		Rate: RateConfig{
			RPS:            getEnvFloat("RATE_LIMIT_RPS", 30),
			Burst:          getEnvInt("RATE_LIMIT_BURST", 60),
			UserMultiplier: getEnvInt("RATE_LIMIT_USER_MULTIPLIER", 10),
		},
		Email: EmailConfig{
			Sender:   os.Getenv("EMAIL_SENDER"),
			Password: os.Getenv("EMAIL_PW"),
			SMTPHost: os.Getenv("EMAIL_SMTP_HOST"),
			SMTPPort: getEnvInt("EMAIL_SMTP_PORT", 0),
		},
		Paths: PathsConfig{
			UploadDir: getEnv("UPLOAD_DIR", "./tmp/uploads/"),
			FilesDir:  getEnv("FILES_DIR", "./files/"),
			PublicDir: getEnv("PUBLIC_DIR", "./public/"),
		},
		Zeitnahme: ZeitnahmeConfig{
			CurrentTag: os.Getenv("ZEITNAHME_CURRENT_TAG"),
		},
		GotenbergURL:   getEnv("GOTENBERG_URL", "http://gotenberg:3000"),
		AppInternalURL: getEnv("APP_INTERNAL_URL", "http://api-dev:8080"),
		AppPublicURL:   getEnv("APP_PUBLIC_URL", "http://localhost:8080"),
		Env:            getEnv("APP_ENV", "prod"),
		Log: LogConfig{
			Level: os.Getenv("LOG_LEVEL"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvRequired(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	slog.Error("required environment variable not set", "key", key)
	return ""
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultVal
}

func getEnvList(key string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}
	return nil
}
