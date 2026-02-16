package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mobileproxy/server/internal/api/handler"
	"github.com/mobileproxy/server/internal/domain"
	"github.com/mobileproxy/server/internal/repository"
	"github.com/mobileproxy/server/internal/service"
)

func main() {
	cfg := loadConfig()

	// Database
	db, err := repository.NewDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	connRepo := repository.NewConnectionRepository(db)
	commandRepo := repository.NewCommandRepository(db)
	ipHistRepo := repository.NewIPHistoryRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	_ = repository.NewBandwidthRepository(db) // Used by worker

	// Services
	portService := service.NewPortService(deviceRepo, cfg.Ports)
	authService := service.NewAuthService(userRepo, cfg.JWT)
	deviceService := service.NewDeviceService(deviceRepo, ipHistRepo, commandRepo, portService)
	connService := service.NewConnectionService(connRepo, deviceRepo)

	// WebSocket hub
	wsHub := handler.NewWSHub()

	// Handlers
	customerHandler := handler.NewCustomerHandler(customerRepo)

	// Router
	router := handler.SetupRouter(authService, deviceService, connService, customerHandler, wsHub)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting API server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadConfig() domain.Config {
	cfg := domain.DefaultConfig()

	// Override from environment variables
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
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}

	return cfg
}
