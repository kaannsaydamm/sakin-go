package handlers

import (
	"github.com/gofiber/fiber/v2"

	"sakin-go/cmd/sge-panel-api/services"
)

type DashboardHandler struct {
	service *services.DashboardService
}

func NewDashboardHandler(s *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: s}
}

func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.service.GetOverview(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(stats)
}
