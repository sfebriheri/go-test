package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"transjakarta-fleet/internal/services"
)

type Handler struct {
	vehicleService *services.VehicleService
}

func NewHandler(vehicleService *services.VehicleService) *Handler {
	return &Handler{
		vehicleService: vehicleService,
	}
}

// GetLastLocation godoc
// @Summary Get last known location of a vehicle
// @Description Retrieves the most recent location data for a specific vehicle
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle_id path string true "Vehicle ID"
// @Success 200 {object} models.VehicleLocation
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vehicles/{vehicle_id}/location [get]
func (h *Handler) GetLastLocation(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")

	location, err := h.vehicleService.GetLastLocation(vehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, location)
}

// GetLocationHistory godoc
// @Summary Get location history of a vehicle
// @Description Retrieves location history for a vehicle within a specified time range
// @Tags vehicles
// @Accept json
// @Produce json
// @Param vehicle_id path string true "Vehicle ID"
// @Param start query int64 true "Start timestamp (Unix epoch)"
// @Param end query int64 true "End timestamp (Unix epoch)"
// @Success 200 {array} models.VehicleLocation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /vehicles/{vehicle_id}/history [get]
func (h *Handler) GetLocationHistory(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")

	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start and end query parameters are required",
		})
		return
	}

	startTime, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid start timestamp",
		})
		return
	}

	endTime, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid end timestamp",
		})
		return
	}

	locations, err := h.vehicleService.GetLocationHistory(vehicleID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if len(locations) == 0 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	c.JSON(http.StatusOK, locations)
}
