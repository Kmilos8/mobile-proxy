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

	// Services
	iptablesService := service.NewIPTablesService()
	vpnService := service.NewVPNService(cfg.VPN, iptablesService)
	portService := service.NewPortService(deviceRepo, cfg.Ports)
	authService := service.NewAuthService(userRepo, cfg.JWT)
	deviceService := service.NewDeviceService(deviceRepo, ipHistRepo, commandRepo, portService, vpnService)
	connService := service.NewConnectionService(connRepo, deviceRepo)
	bwRepo := repository.NewBandwidthRepository(db)
	bwService := service.NewBandwidthService(bwRepo)

	// WebSocket hub
	wsHub := handler.NewWSHub()

	// Handlers
	customerHandler := handler.NewCustomerHandler(customerRepo)
	vpnHandler := handler.NewVPNHandler(deviceService, vpnService)
	statsHandler := handler.NewStatsHandler(deviceRepo, connRepo, bwService)

	// Router
	router := handler.SetupRouter(authService, deviceService, connService, bwService, customerHandler, vpnHandler, statsHandler, wsHub)

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
	if v := os.Getenv("VPN_SERVER_IP"); v != "" {
		cfg.VPN.ServerIP = v
	}
	if v := os.Getenv("VPN_CCD_DIR"); v != "" {
		cfg.VPN.CCDDir = v
	}

	return cfg
}
