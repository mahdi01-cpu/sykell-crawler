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
	urlSerfvice := service.NewURLService(urlRepo)
	srv := httpserver.New(cfg.HTTPAddr, httpserver.Deps{URLService: urlSerfvice})

	// run server in goroutine so we can handle graceful shutdown
	go func() {
		log.Printf("http server listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	} else {
		log.Printf("shutdown complete")
	}
}
