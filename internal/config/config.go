package config

import (
	"os"
)

type Config struct {
	// Database
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string

	// MQTT
	MQTTBroker   string
	MQTTClientID string
	MQTTUsername string
	MQTTPassword string

	// RabbitMQ
	RabbitMQURL      string
	RabbitMQExchange string
	RabbitMQQueue    string

	// Geofence
	GeofenceLatitude  float64
	GeofenceLongitude float64
	GeofenceRadius    float64

	// Server
	ServerPort string
}

func LoadConfig() *Config {
	return &Config{
		// Database
		DatabaseHost:     getEnv("DB_HOST", "localhost"),
		DatabasePort:     getEnv("DB_PORT", "5432"),
		DatabaseUser:     getEnv("DB_USER", "postgres"),
		DatabasePassword: getEnv("DB_PASSWORD", "postgres"),
		DatabaseName:     getEnv("DB_NAME", "transjakarta_fleet"),
		DatabaseSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// MQTT
		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "transjakarta-backend"),
		MQTTUsername: getEnv("MQTT_USERNAME", ""),
		MQTTPassword: getEnv("MQTT_PASSWORD", ""),

		// RabbitMQ
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "fleet.events"),
		RabbitMQQueue:    getEnv("RABBITMQ_QUEUE", "geofence_alerts"),

		// Geofence (Default: Monas, Jakarta)
		GeofenceLatitude:  -6.1751,
		GeofenceLongitude: 106.8270,
		GeofenceRadius:    50.0, // meters

		// Server
		ServerPort: getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
