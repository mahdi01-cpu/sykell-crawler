package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mahdi-01/sykell-crawler/internal/config"
	"github.com/mahdi-01/sykell-crawler/internal/crawler"
	repo "github.com/mahdi-01/sykell-crawler/internal/repo/mysql"
	"github.com/mahdi-01/sykell-crawler/internal/service"
	"github.com/mahdi-01/sykell-crawler/internal/transport/httpserver"
)

func main() {
	// TODO: introduce application struct to hold dependencies and manage lifecycle
	cfg := config.NewConfig()
	sqlDB, err := sql.Open("mysql", cfg.DBDSN)
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer sqlDB.Close()

	urlRepo := repo.NewURLRepo(sqlDB)
	urlService := service.NewURLService(urlRepo)
	srv := httpserver.New(cfg.HTTPAddr, httpserver.Deps{URLService: urlService})

	// root ctx for app lifecycle
	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	go func() {
		log.Printf("http server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	c := crawler.NewCrawler(
		urlRepo,
		cfg.Crawler.ScheduleInterval,
		cfg.Crawler.ScheduleBatchSize,
		cfg.Crawler.WorkerNum,
		cfg.Crawler.ExpirationTimeout,
		cfg.Crawler.CrawlTimeout,
	)
	c.Start(appCtx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("shutting down...")
	cancelApp() // <-- this stops scheduler/workers

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // <-- this is for http server shutdown timeout

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	} else {
		log.Printf("shutdown complete")
	}
}
