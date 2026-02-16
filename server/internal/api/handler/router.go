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
	customerHandler *CustomerHandler,
	wsHub *WSHub,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	authHandler := NewAuthHandler(authService)
	deviceHandler := NewDeviceHandler(deviceService, wsHub)
	connHandler := NewConnectionHandler(connService)

	// Public routes
	r.POST("/api/auth/login", authHandler.Login)

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
		dashboard.GET("/devices", deviceHandler.List)
		dashboard.GET("/devices/:id", deviceHandler.GetByID)
		dashboard.POST("/devices/:id/commands", deviceHandler.SendCommand)
		dashboard.GET("/devices/:id/ip-history", deviceHandler.GetIPHistory)

		dashboard.GET("/connections", connHandler.List)
		dashboard.POST("/connections", connHandler.Create)
		dashboard.GET("/connections/:id", connHandler.GetByID)
		dashboard.PATCH("/connections/:id", connHandler.SetActive)
		dashboard.DELETE("/connections/:id", connHandler.Delete)

		dashboard.GET("/customers", customerHandler.List)
		dashboard.POST("/customers", customerHandler.Create)
		dashboard.GET("/customers/:id", customerHandler.GetByID)
		dashboard.PUT("/customers/:id", customerHandler.Update)
	}

	// WebSocket
	r.GET("/ws", wsHub.HandleWS)

	return r
}
