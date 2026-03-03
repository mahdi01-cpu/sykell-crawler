package config

import (
	"os"
)

type Config struct {
	HTTPAddr string
	DBDSN    string
}

func NewConfig() *Config {
	return &Config{
		HTTPAddr: getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		DBDSN:    buildDBDSN(),
	}
}

func buildDBDSN() string {
	dsn := getEnv("DB_DSN", "")
	if dsn != "" {
		return dsn
	}
	host := getEnv("MYSQL_HOST", "localhost")
	port := getEnv("MYSQL_PORT", "3306")
	user := getEnv("MYSQL_USER", "crawler_user")
	password := getEnv("MYSQL_PASSWORD", "supersecret")
	dbname := getEnv("MYSQL_DATABASE", "crawler")

	if password != "" {
		password = ":" + password
	}

	tail := "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC"
	return user + password + "@tcp(" + host + ":" + port + ")/" + dbname + tail
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
