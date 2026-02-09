package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"transjakarta-fleet/internal/config"
	"transjakarta-fleet/internal/models"
	"transjakarta-fleet/internal/services"
)

type MQTTClient struct {
	client         mqtt.Client
	cfg            *config.Config
	vehicleService *services.VehicleService
}

func NewMQTTClient(cfg *config.Config, vehicleService *services.VehicleService) *MQTTClient {
	return &MQTTClient{
		cfg:            cfg,
		vehicleService: vehicleService,
	}
}

func (m *MQTTClient) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.cfg.MQTTBroker)
	opts.SetClientID(m.cfg.MQTTClientID)
	
	if m.cfg.MQTTUsername != "" {
		opts.SetUsername(m.cfg.MQTTUsername)
		opts.SetPassword(m.cfg.MQTTPassword)
	}

	opts.SetDefaultPublishHandler(m.messageHandler)
	opts.SetOnConnectHandler(m.onConnect)
	opts.SetConnectionLostHandler(m.onConnectionLost)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	m.client = mqtt.NewClient(opts)

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Println("Successfully connected to MQTT broker")
	return nil
}

func (m *MQTTClient) onConnect(client mqtt.Client) {
	log.Println("MQTT client connected, subscribing to topics...")
	
	// Subscribe to all vehicle location topics
	topic := "/fleet/vehicle/+/location"
	if token := client.Subscribe(topic, 1, m.messageHandler); token.Wait() && token.Error() != nil {
		log.Printf("Failed to subscribe to topic %s: %v", topic, token.Error())
	} else {
		log.Printf("Successfully subscribed to topic: %s", topic)
	}
}

func (m *MQTTClient) onConnectionLost(client mqtt.Client, err error) {
	log.Printf("MQTT connection lost: %v", err)
}

func (m *MQTTClient) messageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message on topic %s: %s", msg.Topic(), string(msg.Payload()))

	var location models.VehicleLocation
	if err := json.Unmarshal(msg.Payload(), &location); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	// Validate location data
	if location.VehicleID == "" {
		log.Println("Invalid location data: missing vehicle_id")
		return
	}

	if location.Latitude < -90 || location.Latitude > 90 {
		log.Printf("Invalid latitude: %f", location.Latitude)
		return
	}

	if location.Longitude < -180 || location.Longitude > 180 {
		log.Printf("Invalid longitude: %f", location.Longitude)
		return
	}

	// Save location to database
	if err := m.vehicleService.SaveLocation(&location); err != nil {
		log.Printf("Failed to save location: %v", err)
		return
	}

	log.Printf("Successfully saved location for vehicle %s", location.VehicleID)
}

func (m *MQTTClient) Disconnect() {
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(250)
		log.Println("MQTT client disconnected")
	}
}
