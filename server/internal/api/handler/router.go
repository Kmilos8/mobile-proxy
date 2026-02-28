package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mobileproxy/server/internal/api/middleware"
	"github.com/mobileproxy/server/internal/repository"
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
	syncHandler *SyncHandler,
	userRepo *repository.UserRepository,
	customerAuthHandler *CustomerAuthHandler,
	deviceShareHandler *DeviceShareHandler,
	customerRepo *repository.CustomerRepository,
	shareService *service.DeviceShareService,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	authHandler := NewAuthHandler(authService)
	deviceHandler := NewDeviceHandler(deviceService, bwService, wsHub, shareService)
	deviceHandler.SetConnectionService(connService)
	connHandler := NewConnectionHandler(connService)
	connHandler.SetShareService(shareService)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	r.POST("/api/auth/login", authHandler.Login)

	// Customer auth routes (public — no JWT required)
	customerAuth := r.Group("/api/auth/customer")
	{
		customerAuth.POST("/signup", customerAuthHandler.Signup)
		customerAuth.POST("/login", customerAuthHandler.Login)
		customerAuth.GET("/verify-email", customerAuthHandler.VerifyEmailCheck)
		customerAuth.POST("/verify-email", customerAuthHandler.VerifyEmail)
		customerAuth.POST("/resend-verification", customerAuthHandler.ResendVerification)
		customerAuth.POST("/forgot-password", customerAuthHandler.ForgotPassword)
		customerAuth.POST("/reset-password", customerAuthHandler.ResetPassword)
		customerAuth.GET("/google", customerAuthHandler.GoogleLogin)
		customerAuth.GET("/google/callback", customerAuthHandler.GoogleCallback)
	}

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

	// Dashboard routes (JWT protected) — all authenticated users pass here
	dashboard := r.Group("/api")
	dashboard.Use(middleware.AuthMiddleware(authService))
	dashboard.Use(middleware.CustomerSuspensionCheck(customerRepo))

	// Admin-only sub-group: blocks customer tokens with 403
	adminOnly := dashboard.Group("")
	adminOnly.Use(middleware.AdminOnlyMiddleware())
	{
		adminOnly.GET("/stats/overview", statsHandler.Overview)

		adminOnly.GET("/customers", customerHandler.List)
		adminOnly.POST("/customers", customerHandler.Create)
		adminOnly.GET("/customers/:id", customerHandler.GetDetail)
		adminOnly.PUT("/customers/:id", customerHandler.Update)
		adminOnly.POST("/customers/:id/suspend", customerHandler.Suspend)
		adminOnly.POST("/customers/:id/activate", customerHandler.Activate)

		adminOnly.GET("/rotation-links", rotationLinkHandler.List)
		adminOnly.POST("/rotation-links", rotationLinkHandler.Create)
		adminOnly.DELETE("/rotation-links/:id", rotationLinkHandler.Delete)

		adminOnly.GET("/pairing-codes", pairingHandler.ListCodes)
		adminOnly.POST("/pairing-codes", pairingHandler.CreateCode)
		adminOnly.DELETE("/pairing-codes/:id", pairingHandler.DeleteCode)

		adminOnly.GET("/relay-servers", relayServerHandler.List)
		adminOnly.GET("/relay-servers/active", relayServerHandler.ListActive)
		adminOnly.POST("/relay-servers", relayServerHandler.Create)

		// Settings: webhook URL management (admin only)
		adminOnly.GET("/settings/webhook", func(c *gin.Context) {
			userIDVal, _ := c.Get("user_id")
			uid := userIDVal.(uuid.UUID)
			user, err := userRepo.GetByID(c.Request.Context(), uid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"webhook_url": user.WebhookURL})
		})

		adminOnly.PUT("/settings/webhook", func(c *gin.Context) {
			var body struct {
				WebhookURL string `json:"webhook_url"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			userIDVal, _ := c.Get("user_id")
			uid := userIDVal.(uuid.UUID)
			var urlPtr *string
			if body.WebhookURL != "" {
				urlPtr = &body.WebhookURL
			}
			if err := userRepo.UpdateWebhookURL(c.Request.Context(), uid, urlPtr); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		adminOnly.POST("/settings/webhook/test", func(c *gin.Context) {
			var body struct {
				WebhookURL string `json:"webhook_url"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			payload, _ := json.Marshal(map[string]interface{}{
				"event":     "test",
				"message":   "This is a test webhook from PocketProxy",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Post(body.WebhookURL, "application/json", bytes.NewReader(payload))
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			resp.Body.Close()
			c.JSON(http.StatusOK, gin.H{"ok": true, "status": resp.StatusCode})
		})
	}

	// Mixed-access routes: device and connection endpoints (handlers branch internally by role)
	{
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
		dashboard.POST("/connections/:id/regenerate-password", connHandler.RegeneratePassword)
		dashboard.POST("/connections/:id/reset-bandwidth", connHandler.ResetBandwidth)

		// Device shares (accessible to authenticated users — handler checks ownership)
		dashboard.GET("/device-shares", deviceShareHandler.ListShares)
		dashboard.POST("/device-shares", deviceShareHandler.CreateShare)
		dashboard.PUT("/device-shares/:id", deviceShareHandler.UpdateShare)
		dashboard.DELETE("/device-shares/:id", deviceShareHandler.DeleteShare)
	}

	// Internal VPN routes (called by OpenVPN scripts)
	if vpnHandler != nil {
		internal := r.Group("/api/internal")
		{
			internal.POST("/vpn/connected", vpnHandler.Connected)
			internal.POST("/vpn/disconnected", vpnHandler.Disconnected)
		}
	}

	// Internal sync routes (called by peer server)
	if syncHandler != nil {
		syncGroup := r.Group("/api/internal/sync")
		{
			syncGroup.POST("/device", syncHandler.SyncDevice)
			syncGroup.POST("/connections", syncHandler.SyncConnections)
		}
	}

	// Internal bandwidth flush (called by tunnel server, no JWT)
	r.POST("/api/internal/bandwidth-flush", connHandler.BandwidthFlush)

	// Internal OpenVPN client routes (called by OpenVPN client-server scripts)
	if openvpnHandler != nil {
		ovpnInternal := r.Group("/api/internal/openvpn")
		{
			ovpnInternal.POST("/auth", openvpnHandler.Auth)
			ovpnInternal.POST("/connect", openvpnHandler.Connect)
			ovpnInternal.POST("/disconnect", openvpnHandler.Disconnect)
		}
		// Dashboard route for .ovpn download (JWT protected, customer needs download_configs permission)
		dashboard.GET("/connections/:id/ovpn", openvpnHandler.DownloadOVPN)
	}

	// WebSocket
	r.GET("/ws", wsHub.HandleWS)

	return r
}
