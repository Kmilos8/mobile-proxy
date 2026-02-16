package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
	"github.com/mobileproxy/server/internal/service"
)

func main() {
	cfg := loadConfig()

	db, err := repository.NewDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Worker connected to database")

	deviceRepo := repository.NewDeviceRepository(db)
	ipHistRepo := repository.NewIPHistoryRepository(db)
	commandRepo := repository.NewCommandRepository(db)
	bwRepo := repository.NewBandwidthRepository(db)

	portService := service.NewPortService(deviceRepo, cfg.Ports)
	deviceService := service.NewDeviceService(deviceRepo, ipHistRepo, commandRepo, portService, nil)
	bwService := service.NewBandwidthService(bwRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Stale device checker - every 30 seconds
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, err := deviceService.MarkStaleOffline(ctx)
				if err != nil {
					log.Printf("Error marking stale devices: %v", err)
				} else if count > 0 {
					log.Printf("Marked %d stale devices as offline", count)
				}
			}
		}
	}()

	// Partition maintainer - every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		// Run immediately on start
		if err := bwService.EnsurePartitions(ctx); err != nil {
			log.Printf("Error ensuring partitions: %v", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := bwService.EnsurePartitions(ctx); err != nil {
					log.Printf("Error ensuring partitions: %v", err)
				}
			}
		}
	}()

	log.Println("Worker started")
	<-sigCh
	log.Println("Worker shutting down")
	cancel()
}

func loadConfig() domain.Config {
	cfg := domain.DefaultConfig()
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Database.Port)
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	return cfg
}
