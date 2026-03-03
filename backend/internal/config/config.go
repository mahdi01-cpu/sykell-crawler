package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type CrawlerConfig struct {
	ScheduleInterval  time.Duration
	ScheduleBatchSize int
	WorkerNum         int
	ExpirationTimeout time.Duration
	CrawlTimeout      time.Duration
}
type Config struct {
	HTTPAddr string
	DBDSN    string
	Crawler  *CrawlerConfig
}

func NewConfig() *Config {
	return &Config{
		HTTPAddr: getEnv("HTTP_ADDR", "0.0.0.0:8080"),
		DBDSN:    buildDBDSN(),
		Crawler:  getCrawlerConfig(),
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

func getCrawlerConfig() *CrawlerConfig {
	scheduleIntervalRaw := getEnv("CRAWL_SCHEDULE_INTERVAL", "10s") // default to 10 seconds
	crawlScheduleInterval, err := time.ParseDuration(scheduleIntervalRaw)
	if err != nil {
		fmt.Printf("invalid CRAWL_SCHEDULE_INTERVAL: %v, defaulting to 10s\n", err)
		crawlScheduleInterval = 10 * time.Second
	}

	scheduleBatchSizeRaw := getEnv("CRAWL_SCHEDULE_BATCH_SIZE", "10") // default to 10
	crawlScheduleBatchSize, err := strconv.Atoi(scheduleBatchSizeRaw)
	if err != nil {
		fmt.Printf("invalid CRAWL_SCHEDULE_BATCH_SIZE: %v, defaulting to 10\n", err)
		crawlScheduleBatchSize = 10
	}

	workerNumRaw := getEnv("CRAWL_WORKER_NUM", "5") // default to 5
	crawlWorkerNum, err := strconv.Atoi(workerNumRaw)
	if err != nil {
		fmt.Printf("invalid CRAWL_WORKER_NUM: %v, defaulting to 5\n", err)
		crawlWorkerNum = 5
	}

	expirationTimeoutRaw := getEnv("CRAWL_EXPIRATION_TIMEOUT", "24h")
	crawlExpirationTimeout, err := time.ParseDuration(expirationTimeoutRaw)
	if err != nil {
		fmt.Printf("invalid CRAWL_EXPIRATION_TIMEOUT: %v, defaulting to 24h\n", err)
		crawlExpirationTimeout = 24 * time.Hour
	}

	crawlTimeoutRaw := getEnv("CRAWL_TIMEOUT", "30s")
	crawlTimeout, err := time.ParseDuration(crawlTimeoutRaw)
	if err != nil {
		fmt.Printf("invalid CRAWL_TIMEOUT: %v, defaulting to 30s\n", err)
		crawlTimeout = 30 * time.Second
	}

	return &CrawlerConfig{
		ScheduleInterval:  crawlScheduleInterval,
		ScheduleBatchSize: crawlScheduleBatchSize,
		WorkerNum:         crawlWorkerNum,
		ExpirationTimeout: crawlExpirationTimeout,
		CrawlTimeout:      crawlTimeout,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
