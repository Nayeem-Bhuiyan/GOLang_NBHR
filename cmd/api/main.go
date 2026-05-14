package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"nbhr/config"
	"nbhr/internal/bootstrap"
	"nbhr/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "run database migrations and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if *migrateOnly {
		if err := runMigrations(cfg); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		fmt.Println("migrations completed successfully")
		os.Exit(0)
	}

	app, err := bootstrap.New(cfg)
	if err != nil {
		log.Fatalf("failed to bootstrap application: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func runMigrations(cfg *config.Config) error {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	return migrations.Run(db)
}