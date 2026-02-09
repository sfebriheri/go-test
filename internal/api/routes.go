package api

import (
	"github.com/gin-gonic/gin"
	"transjakarta-fleet/internal/services"
)

func SetupRoutes(router *gin.Engine, vehicleService *services.VehicleService) {
	handler := NewHandler(vehicleService)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		vehicles := v1.Group("/vehicles")
		{
			vehicles.GET("/:vehicle_id/location", handler.GetLastLocation)
			vehicles.GET("/:vehicle_id/history", handler.GetLocationHistory)
		}
	}
}
