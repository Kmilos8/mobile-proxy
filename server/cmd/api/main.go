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
	customerTokenRepo := repository.NewCustomerTokenRepository(db)
	rotationLinkRepo := repository.NewRotationLinkRepository(db)
	pairingRepo := repository.NewPairingCodeRepository(db)
	relayServerRepo := repository.NewRelayServerRepository(db)
	deviceShareRepo := repository.NewDeviceShareRepository(db)

	// Services
	iptablesService := service.NewIPTablesService()
	vpnService := service.NewVPNService(cfg.VPN, iptablesService)
	portService := service.NewPortService(deviceRepo, cfg.Ports)
	authService := service.NewAuthService(userRepo, cfg.JWT)
	statusLogRepo := repository.NewStatusLogRepository(db)
	deviceService := service.NewDeviceService(deviceRepo, ipHistRepo, commandRepo, portService, vpnService)
	deviceService.SetStatusLogRepo(statusLogRepo)
	deviceService.SetUserRepo(userRepo)
	deviceService.SetRelayServerRepo(relayServerRepo)
	if v := os.Getenv("TUNNEL_PUSH_URL"); v != "" {
		deviceService.SetTunnelPushURL(v)
		log.Printf("Tunnel push URL configured: %s", v)
	}
	connService := service.NewConnectionService(connRepo, deviceRepo)
	connService.SetPortService(portService)
	connService.SetRelayServerRepo(relayServerRepo)
	if v := os.Getenv("TUNNEL_PUSH_URL"); v != "" {
		connService.SetTunnelPushURL(v)
	}
	bwRepo := repository.NewBandwidthRepository(db)
	bwService := service.NewBandwidthService(bwRepo)
	relayServerService := service.NewRelayServerService(relayServerRepo)

	// Device share service (multi-tenant permission layer)
	deviceShareService := service.NewDeviceShareService(deviceShareRepo, deviceRepo)

	// Build server URL for pairing responses
	serverURL := fmt.Sprintf("http://%s:%d", cfg.VPN.ServerIP, cfg.Server.Port)
	if v := os.Getenv("PUBLIC_SERVER_URL"); v != "" {
		serverURL = v
	}
	pairingService := service.NewPairingService(pairingRepo, deviceService, deviceRepo, connRepo, relayServerRepo, serverURL)

	// Peer sync service
	var syncService *service.SyncService
	if v := os.Getenv("PEER_API_URL"); v != "" {
		syncService = service.NewSyncService(v)
		pairingService.SetSyncService(syncService)
		connService.SetSyncService(syncService)
		log.Printf("Peer sync configured: %s", v)
	}

	// WebSocket hub
	wsHub := handler.NewWSHub()

	// Customer auth services
	emailService := service.NewEmailService(cfg.Resend)
	customerAuthService := service.NewCustomerAuthService(
		customerRepo, customerTokenRepo, emailService,
		cfg.JWT, cfg.Google, cfg.Turnstile,
	)

	// Handlers
	customerHandler := handler.NewCustomerHandler(customerRepo)
	customerAuthHandler := handler.NewCustomerAuthHandler(customerAuthService)
	vpnHandler := handler.NewVPNHandler(deviceService, vpnService, connService)
	statsHandler := handler.NewStatsHandler(deviceRepo, connRepo, bwService)
	rotationLinkHandler := handler.NewRotationLinkHandler(rotationLinkRepo, deviceService)
	pairingHandler := handler.NewPairingHandler(pairingService)
	relayServerHandler := handler.NewRelayServerHandler(relayServerService)
	openvpnHandler := handler.NewOpenVPNHandler(connRepo, deviceService)
	openvpnHandler.SetShareService(deviceShareService)
	syncHandler := handler.NewSyncHandler(deviceRepo, connRepo)
	deviceShareHandler := handler.NewDeviceShareHandler(deviceShareService)

	// Router
	router := handler.SetupRouter(
		authService, deviceService, connService, bwService,
		customerHandler, vpnHandler, statsHandler, rotationLinkHandler,
		pairingHandler, relayServerHandler, wsHub, openvpnHandler, syncHandler,
		userRepo, customerAuthHandler,
		deviceShareHandler, customerRepo, deviceShareService,
	)

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

	// Google OAuth
	if v := os.Getenv("GOOGLE_CLIENT_ID"); v != "" {
		cfg.Google.ClientID = v
	}
	if v := os.Getenv("GOOGLE_CLIENT_SECRET"); v != "" {
		cfg.Google.ClientSecret = v
	}
	if v := os.Getenv("GOOGLE_REDIRECT_URL"); v != "" {
		cfg.Google.RedirectURL = v
	}

	// Resend email
	if v := os.Getenv("RESEND_API_KEY"); v != "" {
		cfg.Resend.APIKey = v
	}
	if v := os.Getenv("RESEND_FROM_EMAIL"); v != "" {
		cfg.Resend.FromEmail = v
	}
	if v := os.Getenv("DASHBOARD_BASE_URL"); v != "" {
		cfg.Resend.BaseURL = v
	}

	// Cloudflare Turnstile
	if v := os.Getenv("TURNSTILE_SITE_KEY"); v != "" {
		cfg.Turnstile.SiteKey = v
	}
	if v := os.Getenv("TURNSTILE_SECRET_KEY"); v != "" {
		cfg.Turnstile.SecretKey = v
	}

	return cfg
}
