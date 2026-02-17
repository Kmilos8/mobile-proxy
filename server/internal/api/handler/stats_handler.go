package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mobileproxy/server/internal/repository"
	"github.com/mobileproxy/server/internal/service"
)

type StatsHandler struct {
	deviceRepo *repository.DeviceRepository
	connRepo   *repository.ConnectionRepository
	bwService  *service.BandwidthService
}

func NewStatsHandler(
	deviceRepo *repository.DeviceRepository,
	connRepo *repository.ConnectionRepository,
	bwService *service.BandwidthService,
) *StatsHandler {
	return &StatsHandler{
		deviceRepo: deviceRepo,
		connRepo:   connRepo,
		bwService:  bwService,
	}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	ctx := c.Request.Context()

	devicesTotal, devicesOnline, err := h.deviceRepo.CountByStatus(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	connectionsActive, err := h.connRepo.CountActive(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todayIn, todayOut, err := h.bwService.GetTotalToday(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	monthIn, monthOut, err := h.bwService.GetTotalMonth(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"devices_total":      devicesTotal,
		"devices_online":     devicesOnline,
		"connections_active": connectionsActive,
		"bandwidth_today_in":  todayIn,
		"bandwidth_today_out": todayOut,
		"bandwidth_month_in":  monthIn,
		"bandwidth_month_out": monthOut,
	})
}
