package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "transjakarta-fleet/docs"
	"transjakarta-fleet/internal/api"
	"transjakarta-fleet/internal/config"
	"transjakarta-fleet/internal/database"
	"transjakarta-fleet/internal/mqtt"
	"transjakarta-fleet/internal/rabbitmq"
	"transjakarta-fleet/internal/services"
)

// @title Transjakarta Fleet Management API
// @version 1.0
// @description API untuk sistem manajemen armada Transjakarta
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@transjakarta.co.id

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.LoadConfig()

	// Initialize PostgreSQL database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize RabbitMQ
	rabbitConn, err := rabbitmq.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	// Initialize services
	vehicleService := services.NewVehicleService(db, rabbitConn, cfg)

	// Initialize MQTT subscriber
	mqttClient := mqtt.NewMQTTClient(cfg, vehicleService)
	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttClient.Disconnect()

	// Start geofence worker
	go rabbitmq.StartGeofenceWorker(rabbitConn)

	// Initialize Gin router
	router := gin.Default()

	// Setup API routes
	api.SetupRoutes(router, vehicleService)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "transjakarta-fleet-management",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
