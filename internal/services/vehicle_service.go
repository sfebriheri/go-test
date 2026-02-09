package services

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	"transjakarta-fleet/internal/config"
	"transjakarta-fleet/internal/models"
	"transjakarta-fleet/internal/rabbitmq"
)

type VehicleService struct {
	db     *sql.DB
	rabbit *rabbitmq.RabbitMQ
	cfg    *config.Config
}

func NewVehicleService(db *sql.DB, rabbit *rabbitmq.RabbitMQ, cfg *config.Config) *VehicleService {
	return &VehicleService{
		db:     db,
		rabbit: rabbit,
		cfg:    cfg,
	}
}

// SaveLocation saves vehicle location to database and checks geofence
func (s *VehicleService) SaveLocation(location *models.VehicleLocation) error {
	query := `
		INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.Exec(query, location.VehicleID, location.Latitude, location.Longitude, location.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save location: %w", err)
	}

	// Check geofence
	if s.isInsideGeofence(location.Latitude, location.Longitude) {
		event := &models.GeofenceEvent{
			VehicleID: location.VehicleID,
			Event:     "geofence_entry",
			Location: models.Location{
				Latitude:  location.Latitude,
				Longitude: location.Longitude,
			},
			Timestamp: location.Timestamp,
		}

		if err := s.rabbit.PublishGeofenceEvent(event); err != nil {
			log.Printf("Failed to publish geofence event: %v", err)
		}
	}

	return nil
}

// GetLastLocation retrieves the last known location of a vehicle
func (s *VehicleService) GetLastLocation(vehicleID string) (*models.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	location := &models.VehicleLocation{}
	err := s.db.QueryRow(query, vehicleID).Scan(
		&location.VehicleID,
		&location.Latitude,
		&location.Longitude,
		&location.Timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no location found for vehicle %s", vehicleID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	return location, nil
}

// GetLocationHistory retrieves location history for a vehicle within a time range
func (s *VehicleService) GetLocationHistory(vehicleID string, startTime, endTime int64) ([]*models.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, vehicleID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query location history: %w", err)
	}
	defer rows.Close()

	var locations []*models.VehicleLocation
	for rows.Next() {
		location := &models.VehicleLocation{}
		if err := rows.Scan(
			&location.VehicleID,
			&location.Latitude,
			&location.Longitude,
			&location.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		locations = append(locations, location)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return locations, nil
}

// isInsideGeofence checks if coordinates are within the geofence radius
func (s *VehicleService) isInsideGeofence(lat, lon float64) bool {
	distance := s.haversineDistance(
		s.cfg.GeofenceLatitude,
		s.cfg.GeofenceLongitude,
		lat,
		lon,
	)
	return distance <= s.cfg.GeofenceRadius
}

// haversineDistance calculates the distance between two points in meters
func (s *VehicleService) haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
