package config

import (
	"os"
	"strconv"
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
	Port   string
	Host   string
	WSPort string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
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
			Port:   getEnv("PORT", "3000"),
			Host:   getEnv("HOST", "127.0.0.1"),
			WSPort: getEnv("WS_PORT", "8081"),
		},
		DB: DBConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "DO_NOT_USE_IN_PROD"),
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
