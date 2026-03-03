package main

import (
	"log"

	"github.com/mahdi-01/sykell-crawler/internal/config"
	"github.com/mahdi-01/sykell-crawler/internal/db"
)

func main() {
	cfg := config.NewConfig()
	if err := db.RunMigrations(cfg.DBDSN); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied")
}
