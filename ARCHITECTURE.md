# Architecture Documentation - Transjakarta Fleet Management System

## System Overview

Sistem Fleet Management Transjakarta adalah aplikasi backend berbasis microservices yang mengintegrasikan multiple technologies untuk real-time vehicle tracking, data persistence, dan event-driven geofencing.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     DOCKER COMPOSE NETWORK                       │
│                                                                   │
│  ┌──────────────┐         ┌──────────────┐                      │
│  │   Vehicle    │  MQTT   │  Mosquitto   │                      │
│  │  Simulator   │────────>│     Broker   │                      │
│  │ (Publisher)  │  Topic  │   (1883)     │                      │
│  └──────────────┘         └──────┬───────┘                      │
│                                   │                               │
│                                   │ Subscribe                     │
│                                   │ /fleet/vehicle/+/location     │
│                                   ↓                               │
│                          ┌─────────────────┐                     │
│                          │   Backend API   │                     │
│                          │   (Golang/Gin)  │                     │
│                          │    Port 8080    │                     │
│                          └────┬──────┬─────┘                     │
│                               │      │                            │
│                    Save       │      │    Publish                 │
│                    Location   │      │    Event                   │
│                               ↓      ↓                            │
│                    ┌──────────────┐ ┌──────────────┐            │
│                    │ PostgreSQL   │ │  RabbitMQ    │            │
│                    │   Database   │ │  Exchange    │            │
│                    │   (5432)     │ │  (5672)      │            │
│                    │              │ │              │            │
│                    │ vehicle_     │ │ fleet.events │            │
│                    │ locations    │ │              │            │
│                    └──────────────┘ └──────┬───────┘            │
│                                            │                      │
│                                            │ Queue                │
│                                            │ geofence_alerts      │
│                                            ↓                      │
│                                    ┌──────────────┐              │
│                                    │  Geofence    │              │
│                                    │   Worker     │              │
│                                    │  (Consumer)  │              │
│                                    └──────────────┘              │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Vehicle Simulator (MQTT Publisher)

**Technology:** Go
**File:** `cmd/publisher/main.go`

**Responsibilities:**
- Generate mock GPS data for 3 vehicles
- Publish location data every 2 seconds
- Simulate realistic vehicle movement

**MQTT Topics:**
- `/fleet/vehicle/B1234XYZ/location`
- `/fleet/vehicle/B5678ABC/location`
- `/fleet/vehicle/B9012DEF/location`

**Data Format:**
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

**Starting Positions:**
- B1234XYZ: Near Monas (-6.2088, 106.8456)
- B5678ABC: Monas (inside geofence) (-6.1751, 106.8270)
- B9012DEF: South Jakarta (-6.2297, 106.8186)

---

### 2. MQTT Broker (Eclipse Mosquitto)

**Technology:** Eclipse Mosquitto 2.0
**Port:** 1883 (MQTT), 9001 (WebSocket)

**Responsibilities:**
- Receive location messages from vehicles
- Deliver messages to subscribers (backend)
- Handle connection management

**Configuration:**
- Anonymous connections allowed
- Persistent messages enabled
- Log level: All

---

### 3. Backend Application (Go + Gin)

**Technology:** Golang 1.21 + Gin Framework
**Port:** 8080
**Files:** `main.go`, `internal/*`

#### 3.1 MQTT Subscriber Module

**File:** `internal/mqtt/client.go`

**Responsibilities:**
- Connect to MQTT broker
- Subscribe to vehicle topics
- Parse and validate incoming messages
- Pass data to service layer

**Features:**
- Auto-reconnect on connection loss
- Message validation (lat/lon bounds check)
- Error handling and logging

#### 3.2 Vehicle Service

**File:** `internal/services/vehicle_service.go`

**Responsibilities:**
- Save location data to database
- Calculate geofence proximity
- Trigger geofence events
- Provide business logic

**Geofence Algorithm:**
- Uses Haversine formula for distance calculation
- Earth radius: 6,371,000 meters
- Triggers event when distance ≤ 50 meters

**Formula:**
```go
a = sin²(Δlat/2) + cos(lat1) × cos(lat2) × sin²(Δlon/2)
c = 2 × atan2(√a, √(1−a))
distance = R × c
```

#### 3.3 API Handlers

**File:** `internal/api/handlers.go`

**Endpoints:**

1. **GET /health**
   - Health check endpoint
   - Returns service status

2. **GET /api/v1/vehicles/{vehicle_id}/location**
   - Get last known location
   - Returns most recent location record

3. **GET /api/v1/vehicles/{vehicle_id}/history**
   - Get location history
   - Query parameters: start, end (Unix timestamps)
   - Returns array of locations

**Features:**
- Input validation
- Error handling
- JSON responses
- Swagger documentation

---

### 4. PostgreSQL Database

**Technology:** PostgreSQL 15
**Port:** 5432

**Database Schema:**

```sql
CREATE TABLE vehicle_locations (
    id SERIAL PRIMARY KEY,
    vehicle_id VARCHAR(50) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    timestamp BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vehicle_id ON vehicle_locations(vehicle_id);
CREATE INDEX idx_timestamp ON vehicle_locations(timestamp);
CREATE INDEX idx_vehicle_timestamp ON vehicle_locations(vehicle_id, timestamp DESC);
```

**Indexes:**
- `vehicle_id`: For fast filtering by vehicle
- `timestamp`: For time-range queries
- `vehicle_id + timestamp`: Composite index for history queries

**Performance:**
- Optimized for write-heavy workload
- Indexes for fast reads
- Auto-increment primary key

---

### 5. RabbitMQ Message Broker

**Technology:** RabbitMQ 3 with Management Plugin
**Ports:** 5672 (AMQP), 15672 (Management UI)

**Configuration:**

**Exchange:**
- Name: `fleet.events`
- Type: `topic`
- Durable: Yes

**Queue:**
- Name: `geofence_alerts`
- Durable: Yes
- Routing Key: `geofence.#`

**Message Flow:**
1. Backend publishes to exchange with routing key `geofence.entry`
2. Message routed to `geofence_alerts` queue
3. Worker consumes from queue
4. Worker processes geofence event

**Event Format:**
```json
{
  "vehicle_id": "B5678ABC",
  "event": "geofence_entry",
  "location": {
    "latitude": -6.1751,
    "longitude": 106.8270
  },
  "timestamp": 1715003456
}
```

---

### 6. Geofence Worker

**File:** `internal/rabbitmq/rabbitmq.go` (function: StartGeofenceWorker)

**Responsibilities:**
- Consume messages from geofence_alerts queue
- Process geofence entry events
- Log events (can be extended for notifications, etc.)

**Future Extensions:**
- Send SMS/email notifications
- Trigger webhooks
- Update external systems
- Create alerts in monitoring system

---

## Data Flow

### 1. Location Update Flow

```
Vehicle Simulator → MQTT Broker → Backend MQTT Client
                                         ↓
                                  Validate Data
                                         ↓
                                  Service Layer
                                    ↙        ↘
                            Save to DB    Check Geofence
                                              ↓
                                         Within radius?
                                              ↓ Yes
                                         Publish to RabbitMQ
                                              ↓
                                         Worker Consumes
                                              ↓
                                         Process Event
```

### 2. API Query Flow

```
HTTP Request → Gin Router → Handler
                                ↓
                          Validate Input
                                ↓
                          Service Layer
                                ↓
                          Query Database
                                ↓
                          Format Response
                                ↓
                          Return JSON
```

## Technology Stack Summary

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Backend | Go + Gin | 1.21 | REST API & Business Logic |
| MQTT Broker | Mosquitto | 2.0 | Message Broker |
| Database | PostgreSQL | 15 | Data Persistence |
| Message Queue | RabbitMQ | 3 | Event Processing |
| API Docs | Swagger | 2.0 | Documentation |
| Container | Docker | Latest | Deployment |

## Design Patterns Used

### 1. Publisher-Subscriber Pattern
- MQTT for real-time location updates
- Decouples vehicle simulators from backend

### 2. Repository Pattern
- Service layer abstracts database operations
- Clean separation of concerns

### 3. Event-Driven Architecture
- RabbitMQ for geofence events
- Asynchronous processing
- Scalable event handling

### 4. Microservices Architecture
- Independent, containerized services
- Easy to scale and maintain
- Fault isolation

## Scalability Considerations

### Horizontal Scaling

1. **Backend API:**
   - Can run multiple instances behind load balancer
   - Stateless design enables easy scaling

2. **MQTT Broker:**
   - Can use Mosquitto clustering
   - Or use managed MQTT service (AWS IoT, Azure IoT Hub)

3. **Database:**
   - PostgreSQL replication (master-slave)
   - Partitioning by vehicle_id or timestamp
   - Use TimescaleDB for time-series optimization

4. **RabbitMQ:**
   - Can form clusters
   - Multiple workers for parallel processing

### Vertical Scaling

- Increase container resources in docker-compose.yml
- Tune PostgreSQL parameters (shared_buffers, work_mem)
- Optimize indexes for query patterns

## Security Considerations

### Current Implementation (Development)
- Anonymous MQTT connections
- No API authentication
- Default RabbitMQ credentials

### Production Recommendations

1. **MQTT Security:**
   - Enable username/password authentication
   - Use TLS/SSL encryption
   - Certificate-based authentication

2. **API Security:**
   - Implement JWT authentication
   - API key for different clients
   - Rate limiting

3. **Database Security:**
   - Strong passwords
   - SSL connections
   - Network isolation

4. **RabbitMQ Security:**
   - Change default credentials
   - Use vhosts for isolation
   - Enable SSL

5. **Network Security:**
   - Use private Docker networks
   - Firewall rules
   - VPN for external access

## Monitoring & Observability

### Logging

Current implementation logs:
- MQTT message reception
- Database operations
- Geofence events
- API requests (via Gin middleware)

### Metrics to Monitor

1. **Application Metrics:**
   - MQTT messages/second
   - API request rate
   - Database query performance
   - Geofence event rate

2. **System Metrics:**
   - CPU usage per container
   - Memory usage
   - Network I/O
   - Disk usage

3. **Business Metrics:**
   - Active vehicles
   - Locations tracked
   - Geofence violations
   - API response times

### Recommended Tools

- **Prometheus + Grafana:** Metrics visualization
- **ELK Stack:** Centralized logging
- **Jaeger:** Distributed tracing
- **Sentry:** Error tracking

## Future Enhancements

### Phase 2 Features

1. **Advanced Geofencing:**
   - Multiple geofence zones
   - Polygon geofences (not just circular)
   - Entry/exit events
   - Dwell time calculation

2. **Real-time Dashboard:**
   - WebSocket for live updates
   - Map visualization
   - Vehicle status indicators

3. **Analytics:**
   - Route analysis
   - Speed calculations
   - Idle time detection
   - Fuel efficiency metrics

4. **Alerting:**
   - Email/SMS notifications
   - Webhook integrations
   - Alert rules engine

5. **Mobile App:**
   - Driver app for status updates
   - Admin app for monitoring

### Performance Optimizations

1. **Caching:**
   - Redis for last known locations
   - Reduce database queries

2. **Database:**
   - Partition by date
   - Archive old data
   - Use read replicas

3. **MQTT:**
   - QoS optimization
   - Message batching
   - Compression

## Testing Strategy

### Unit Tests
- Service layer business logic
- Geofence calculations
- Data validation

### Integration Tests
- MQTT message flow
- Database operations
- API endpoints

### End-to-End Tests
- Complete flow from MQTT to API
- Geofence triggering
- Error handling

### Performance Tests
- Load testing API endpoints
- MQTT message throughput
- Database query performance

## Deployment Strategies

### Development
- Docker Compose (current)
- Local development

### Staging
- Kubernetes cluster
- Separate environment
- Production-like setup

### Production
- Kubernetes with auto-scaling
- Managed services (RDS, ElastiCache)
- Multi-region deployment
- Blue-green deployments

## Disaster Recovery

### Backup Strategy
- Database: Daily automated backups
- Configuration: Version controlled
- Logs: Centralized storage

### Recovery Procedures
- Database restore from backup
- Container restart automation
- Failover procedures

## Conclusion

This architecture provides a solid foundation for a scalable, maintainable fleet management system. The use of proven technologies and design patterns ensures reliability while maintaining flexibility for future enhancements.
