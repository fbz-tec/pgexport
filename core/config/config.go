package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	DefaultDBHost   = "localhost"
	DefaultDBPort   = 5432
	DefaultDBUser   = "postgres"
	DefaultDBName   = "postgres"
	DefaultDBDriver = "postgres"
)

type Config struct {
	DBDriver string
	DBUser   string
	DBPass   string
	DBHost   string
	DBPort   int
	DBName   string
	SSLMode  string
}

func LoadConfig() Config {

	_ = godotenv.Load()

	return Config{
		DBDriver: getEnvOrDefault("DB_DRIVER", DefaultDBDriver),
		DBUser:   getEnvOrDefault("DB_USER", DefaultDBUser),
		DBPass:   os.Getenv("DB_PASS"),
		DBHost:   getEnvOrDefault("DB_HOST", DefaultDBHost),
		DBPort:   getEnvOrDefaultInt("DB_PORT", DefaultDBPort),
		DBName:   getEnvOrDefault("DB_NAME", DefaultDBName),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
}

func (c Config) Validate() error {

	if c.DBPort < 1 || c.DBPort > 65535 {
		return fmt.Errorf("DB_PORT must be a valid port number (1-65535)")
	}

	if strings.TrimSpace(c.DBHost) == "" {
		return fmt.Errorf("DB_HOST cannot be empty or contain only whitespace")
	}

	if strings.TrimSpace(c.DBName) == "" {
		return fmt.Errorf("DB_NAME cannot be empty or contain only whitespace")
	}

	if strings.TrimSpace(c.DBUser) == "" {
		return fmt.Errorf("DB_USER cannot be empty or contain only whitespace")
	}

	return nil
}

func (c Config) GetConnectionString() string {
	u := &url.URL{
		Scheme: c.DBDriver,
		User:   url.UserPassword(c.DBUser, c.DBPass),
		Host:   fmt.Sprintf("%s:%d", c.DBHost, c.DBPort),
		Path:   c.DBName,
	}
	q := u.Query()
	if strings.TrimSpace(c.SSLMode) != "" {
		q.Set("sslmode", c.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		p, err := strconv.Atoi(value)
		if err == nil {
			return p
		}
	}
	return defaultValue
}
