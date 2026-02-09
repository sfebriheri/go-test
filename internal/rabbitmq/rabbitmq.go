package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"transjakarta-fleet/internal/config"
	"transjakarta-fleet/internal/models"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.Config
}

func NewRabbitMQ(cfg *config.Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		cfg.RabbitMQExchange, // name
		"topic",              // type
		true,                 // durable
		false,                // auto-deleted
		false,                // internal
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		cfg.RabbitMQQueue, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		cfg.RabbitMQQueue,    // queue name
		"geofence.#",         // routing key
		cfg.RabbitMQExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		cfg:     cfg,
	}, nil
}

func (r *RabbitMQ) PublishGeofenceEvent(event *models.GeofenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		r.cfg.RabbitMQExchange, // exchange
		"geofence.entry",        // routing key
		false,                   // mandatory
		false,                   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published geofence event for vehicle %s", event.VehicleID)
	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// StartGeofenceWorker starts a worker to consume geofence events
func StartGeofenceWorker(rabbit *RabbitMQ) {
	msgs, err := rabbit.channel.Consume(
		rabbit.cfg.RabbitMQQueue, // queue
		"",                        // consumer
		true,                      // auto-ack
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	log.Println("Geofence worker started, waiting for messages...")

	for msg := range msgs {
		var event models.GeofenceEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal geofence event: %v", err)
			continue
		}

		log.Printf("Received geofence event: Vehicle %s entered geofence at (%.6f, %.6f) at timestamp %d",
			event.VehicleID,
			event.Location.Latitude,
			event.Location.Longitude,
			event.Timestamp,
		)

		// Here you can add additional processing logic
		// For example: send notification, update database, trigger alerts, etc.
	}
}
