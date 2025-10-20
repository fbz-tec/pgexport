package main

import (
	"os"
)

type Config struct {
	DBDriver string
	DBUser   string
	DBPass   string
	DBHost   string
	DBPort   string
	DBName   string
}

func LoadConfig() Config {
	return Config{
		DBDriver: os.Getenv("DB_DRIVER"),
		DBUser:   os.Getenv("DB_USER"),
		DBPass:   os.Getenv("DB_PASS"),
		DBHost:   os.Getenv("DB_HOST"),
		DBPort:   os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
	}
}
