package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mobileproxy/server/internal/api/middleware"
	"github.com/mobileproxy/server/internal/service"
)

func SetupRouter(
	authService *service.AuthService,
	deviceService *service.DeviceService,
	connService *service.ConnectionService,
	bwService *service.BandwidthService,
	customerHandler *CustomerHandler,
	vpnHandler *VPNHandler,
	statsHandler *StatsHandler,
	rotationLinkHandler *RotationLinkHandler,
	pairingHandler *PairingHandler,
	relayServerHandler *RelayServerHandler,
	wsHub *WSHub,
	openvpnHandler *OpenVPNHandler,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	authHandler := NewAuthHandler(authService)
	deviceHandler := NewDeviceHandler(deviceService, bwService, wsHub)
	deviceHandler.SetConnectionService(connService)
	connHandler := NewConnectionHandler(connService)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	r.POST("/api/auth/login", authHandler.Login)

	// Public endpoints (no auth)
	r.GET("/api/public/rotate/:token", rotationLinkHandler.Rotate)
	r.POST("/api/public/pair", pairingHandler.ClaimCode)

	// Device routes (authenticated by device token in future, open for MVP)
	deviceAPI := r.Group("/api/devices")
	{
		deviceAPI.POST("/register", deviceHandler.Register)
		deviceAPI.POST("/:id/heartbeat", deviceHandler.Heartbeat)
		deviceAPI.POST("/:id/commands/:commandId/result", deviceHandler.CommandResult)
	}

	// Dashboard routes (JWT protected)
	dashboard := r.Group("/api")
	dashboard.Use(middleware.AuthMiddleware(authService))
	{
		dashboard.GET("/stats/overview", statsHandler.Overview)

		dashboard.GET("/devices", deviceHandler.List)
		dashboard.GET("/devices/:id", deviceHandler.GetByID)
		dashboard.PATCH("/devices/:id", deviceHandler.Update)
		dashboard.POST("/devices/:id/commands", deviceHandler.SendCommand)
		dashboard.GET("/devices/:id/ip-history", deviceHandler.GetIPHistory)
		dashboard.GET("/devices/:id/bandwidth", deviceHandler.GetBandwidth)
		dashboard.GET("/devices/:id/bandwidth/hourly", deviceHandler.GetBandwidthHourly)
		dashboard.GET("/devices/:id/uptime", deviceHandler.GetUptime)
		dashboard.GET("/devices/:id/commands", deviceHandler.GetCommands)

		dashboard.GET("/connections", connHandler.List)
		dashboard.POST("/connections", connHandler.Create)
		dashboard.GET("/connections/:id", connHandler.GetByID)
		dashboard.PATCH("/connections/:id", connHandler.SetActive)
		dashboard.DELETE("/connections/:id", connHandler.Delete)

		dashboard.GET("/customers", customerHandler.List)
		dashboard.POST("/customers", customerHandler.Create)
		dashboard.GET("/customers/:id", customerHandler.GetByID)
		dashboard.PUT("/customers/:id", customerHandler.Update)

		dashboard.GET("/rotation-links", rotationLinkHandler.List)
		dashboard.POST("/rotation-links", rotationLinkHandler.Create)
		dashboard.DELETE("/rotation-links/:id", rotationLinkHandler.Delete)

		dashboard.GET("/pairing-codes", pairingHandler.ListCodes)
		dashboard.POST("/pairing-codes", pairingHandler.CreateCode)
		dashboard.DELETE("/pairing-codes/:id", pairingHandler.DeleteCode)

		dashboard.GET("/relay-servers", relayServerHandler.List)
		dashboard.GET("/relay-servers/active", relayServerHandler.ListActive)
		dashboard.POST("/relay-servers", relayServerHandler.Create)
	}

	// Internal VPN routes (called by OpenVPN scripts)
	if vpnHandler != nil {
		internal := r.Group("/api/internal")
		{
			internal.POST("/vpn/connected", vpnHandler.Connected)
			internal.POST("/vpn/disconnected", vpnHandler.Disconnected)
		}
	}

	// Internal OpenVPN client routes (called by OpenVPN client-server scripts)
	if openvpnHandler != nil {
		ovpnInternal := r.Group("/api/internal/openvpn")
		{
			ovpnInternal.POST("/auth", openvpnHandler.Auth)
			ovpnInternal.POST("/connect", openvpnHandler.Connect)
			ovpnInternal.POST("/disconnect", openvpnHandler.Disconnect)
		}
		// Dashboard route for .ovpn download (JWT protected)
		dashboard.GET("/connections/:id/ovpn", openvpnHandler.DownloadOVPN)
	}

	// WebSocket
	r.GET("/ws", wsHub.HandleWS)

	return r
}
