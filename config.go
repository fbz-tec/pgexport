package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	DefaultDBHost   = "localhost"
	DefaultDBPort   = "5432"
	DefaultDBUser   = "postgres"
	DefaultDBName   = "postgres"
	DefaultDBDriver = "postgres"
)

type Config struct {
	DBDriver string
	DBUser   string
	DBPass   string
	DBHost   string
	DBPort   string
	DBName   string
	SSLMode  string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}
	return Config{
		DBDriver: getEnvOrDefault("DB_DRIVER", DefaultDBDriver),
		DBUser:   getEnvOrDefault("DB_USER", DefaultDBUser),
		DBPass:   os.Getenv("DB_PASS"),
		DBHost:   getEnvOrDefault("DB_HOST", DefaultDBHost),
		DBPort:   getEnvOrDefault("DB_PORT", DefaultDBPort),
		DBName:   getEnvOrDefault("DB_NAME", DefaultDBName),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
}

func (c Config) Validate() error {

	if port, err := strconv.Atoi(c.DBPort); err != nil || port < 1 || port > 65535 {
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

	if c.DBPass == "" {
		log.Println("Warning: DB_PASS is empty. Connection might fail if password authentication is required.")
	}

	return nil
}

func (c Config) GetConnectionString() string {
	u := &url.URL{
		Scheme: c.DBDriver,
		User:   url.UserPassword(c.DBUser, c.DBPass),
		Host:   fmt.Sprintf("%s:%s", c.DBHost, c.DBPort),
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
