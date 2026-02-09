# Transjakarta Fleet Management System

Sistem Backend untuk manajemen armada Transjakarta yang mengintegrasikan MQTT, PostgreSQL, RabbitMQ, dan Docker.

## 🚀 Fitur Utama

- **MQTT Integration**: Menerima data lokasi kendaraan real-time melalui Eclipse Mosquitto
- **PostgreSQL Database**: Penyimpanan data lokasi kendaraan yang efisien
- **REST API**: Endpoint untuk mengakses lokasi terakhir dan riwayat perjalanan
- **Geofencing**: Deteksi otomatis ketika kendaraan memasuki area tertentu (radius 50 meter)
- **RabbitMQ Events**: Event-driven architecture untuk notifikasi geofence
- **Swagger Documentation**: API documentation yang lengkap dan interaktif
- **Docker Support**: Containerized deployment untuk semua komponen

## 📋 Teknologi yang Digunakan

- **Backend**: Golang 1.21 dengan Gin framework
- **MQTT Broker**: Eclipse Mosquitto 2.0
- **Database**: PostgreSQL 15
- **Message Queue**: RabbitMQ 3 with Management Plugin
- **API Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose

## 🏗️ Arsitektur Sistem

```
┌─────────────┐         MQTT          ┌──────────────┐
│   Vehicle   │ ──────────────────────>│  Mosquitto   │
│  Simulator  │   /fleet/vehicle/+/    │     MQTT     │
└─────────────┘      location          └──────┬───────┘
                                              │
                                              │ Subscribe
                                              ↓
                                       ┌──────────────┐
                                       │    Backend   │
                                       │   (Golang)   │
                                       └──┬────────┬──┘
                                          │        │
                            Save Location │        │ Publish Event
                                          ↓        ↓
                                   ┌──────────┐ ┌─────────┐
                                   │PostgreSQL│ │RabbitMQ │
                                   │          │ │ Queue   │
                                   └──────────┘ └────┬────┘
                                                     │
                                                     │ Consume
                                                     ↓
                                              ┌──────────────┐
                                              │   Geofence   │
                                              │    Worker    │
                                              └──────────────┘
```

## 📦 Struktur Proyek

```
transjakarta-fleet-management/
├── cmd/
│   └── publisher/          # MQTT publisher untuk simulasi kendaraan
│       └── main.go
├── internal/
│   ├── api/               # HTTP handlers dan routes
│   │   ├── handlers.go
│   │   └── routes.go
│   ├── config/            # Konfigurasi aplikasi
│   │   └── config.go
│   ├── database/          # Database connection dan migrasi
│   │   └── postgres.go
│   ├── models/            # Data models
│   │   └── vehicle.go
│   ├── mqtt/              # MQTT client dan subscriber
│   │   └── client.go
│   ├── rabbitmq/          # RabbitMQ connection dan publisher
│   │   └── rabbitmq.go
│   └── services/          # Business logic
│       └── vehicle_service.go
├── mosquitto/
│   └── config/
│       └── mosquitto.conf
├── docs/                  # Swagger documentation (auto-generated)
├── .env                   # Environment variables
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile            # Docker image definition
├── go.mod                # Go dependencies
├── go.sum
├── main.go               # Application entry point
└── README.md
```

## 🔧 Instalasi dan Setup

### Prerequisites

- Docker dan Docker Compose
- Git

### Langkah-langkah Instalasi

1. **Clone Repository**
   ```bash
   git clone <repository-url>
   cd transjakarta-fleet-management
   ```

2. **Konfigurasi Environment Variables** (Opsional)
   
   File `.env` sudah dikonfigurasi dengan nilai default. Anda dapat mengubahnya sesuai kebutuhan:
   ```bash
   # Edit .env file jika diperlukan
   nano .env
   ```

3. **Build dan Jalankan Semua Services**
   ```bash
   docker-compose up --build
   ```

   Perintah ini akan:
   - Build Docker image untuk backend dan publisher
   - Start PostgreSQL database
   - Start RabbitMQ dengan management console
   - Start Mosquitto MQTT broker
   - Start backend application
   - Start vehicle simulator (publisher)

4. **Verifikasi Services Berjalan**
   ```bash
   docker-compose ps
   ```

   Semua services harus dalam status "Up".

## 🧪 Testing API

### Mengakses Swagger Documentation

Buka browser dan akses:
```
http://localhost:8080/swagger/index.html
```

### API Endpoints

#### 1. Health Check
```bash
curl http://localhost:8080/health
```

#### 2. Get Last Location
```bash
curl http://localhost:8080/api/v1/vehicles/B1234XYZ/location
```

**Response:**
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

#### 3. Get Location History
```bash
curl "http://localhost:8080/api/v1/vehicles/B1234XYZ/history?start=1715000000&end=1715009999"
```

**Response:**
```json
[
  {
    "vehicle_id": "B1234XYZ",
    "latitude": -6.2088,
    "longitude": 106.8456,
    "timestamp": 1715003456
  },
  {
    "vehicle_id": "B1234XYZ",
    "latitude": -6.2090,
    "longitude": 106.8458,
    "timestamp": 1715003458
  }
]
```

## 📊 Monitoring Services

### 1. RabbitMQ Management Console
```
URL: http://localhost:15672
Username: guest
Password: guest
```

Di sini Anda dapat:
- Melihat queue `geofence_alerts`
- Monitor message rate
- Melihat geofence events yang dikirim

### 2. PostgreSQL Database
```bash
docker exec -it transjakarta-postgres psql -U postgres -d transjakarta_fleet
```

Query untuk melihat data:
```sql
-- Melihat semua lokasi
SELECT * FROM vehicle_locations ORDER BY timestamp DESC LIMIT 10;

-- Melihat lokasi per kendaraan
SELECT * FROM vehicle_locations WHERE vehicle_id = 'B1234XYZ' ORDER BY timestamp DESC;
```

### 3. MQTT Messages
```bash
# Subscribe ke semua topik vehicle
docker exec -it transjakarta-mosquitto mosquitto_sub -t "/fleet/vehicle/+/location" -v
```

### 4. Logs
```bash
# Backend logs
docker logs -f transjakarta-backend

# Publisher logs
docker logs -f transjakarta-publisher

# RabbitMQ logs
docker logs -f transjakarta-rabbitmq

# PostgreSQL logs
docker logs -f transjakarta-postgres
```

## 🎯 Cara Kerja Geofencing

1. **Konfigurasi Geofence** di `.env`:
   ```
   GEOFENCE_LATITUDE=-6.1751    # Monas, Jakarta
   GEOFENCE_LONGITUDE=106.8270
   GEOFENCE_RADIUS=50           # 50 meters
   ```

2. **Deteksi Geofence**:
   - Setiap kali lokasi kendaraan diterima via MQTT
   - Backend menghitung jarak menggunakan Haversine formula
   - Jika jarak ≤ 50 meter dari titik geofence
   - Event dikirim ke RabbitMQ queue `geofence_alerts`

3. **Event Format**:
   ```json
   {
     "vehicle_id": "B1234XYZ",
     "event": "geofence_entry",
     "location": {
       "latitude": -6.1751,
       "longitude": 106.8270
     },
     "timestamp": 1715003456
   }
   ```

## 🔄 Vehicle Simulator

Publisher secara otomatis mengirim data lokasi untuk 3 kendaraan setiap 2 detik:
- `B1234XYZ` - Mulai dekat Monas
- `B5678ABC` - Mulai di Monas (dalam geofence)
- `B9012DEF` - Mulai di Jakarta Selatan

Kendaraan bergerak secara random dengan increment kecil untuk simulasi pergerakan real.

## 🛠️ Development

### Generate Swagger Documentation
```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init
```

### Run Without Docker (Development)
```bash
# Start PostgreSQL, RabbitMQ, dan Mosquitto
docker-compose up postgres rabbitmq mosquitto

# Update .env untuk localhost
# DB_HOST=localhost
# MQTT_BROKER=tcp://localhost:1883
# RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# Run backend
go run main.go

# Run publisher di terminal lain
go run cmd/publisher/main.go
```

## 🧹 Cleanup

```bash
# Stop semua services
docker-compose down

# Stop dan hapus volumes (database data akan terhapus)
docker-compose down -v

# Remove images
docker-compose down --rmi all
```

## 📝 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | postgres | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | Database user |
| DB_PASSWORD | postgres | Database password |
| DB_NAME | transjakarta_fleet | Database name |
| MQTT_BROKER | tcp://mosquitto:1883 | MQTT broker URL |
| RABBITMQ_URL | amqp://guest:guest@rabbitmq:5672/ | RabbitMQ connection URL |
| RABBITMQ_EXCHANGE | fleet.events | RabbitMQ exchange name |
| RABBITMQ_QUEUE | geofence_alerts | RabbitMQ queue name |
| PORT | 8080 | HTTP server port |
| GEOFENCE_LATITUDE | -6.1751 | Geofence center latitude |
| GEOFENCE_LONGITUDE | 106.8270 | Geofence center longitude |
| GEOFENCE_RADIUS | 50 | Geofence radius in meters |

## 🐛 Troubleshooting

### Service tidak bisa connect
```bash
# Restart services
docker-compose restart

# Cek logs
docker-compose logs -f
```

### Database connection error
```bash
# Pastikan PostgreSQL sudah ready
docker-compose ps postgres
docker logs transjakarta-postgres
```

### MQTT connection failed
```bash
# Test MQTT broker
docker exec -it transjakarta-mosquitto mosquitto_sub -t "#" -v
```

## 📞 Support

Untuk pertanyaan dan support, silakan buat issue di repository ini.

## 📄 License

Apache 2.0 License

---

**Dibuat untuk Tes Teknis Backend Engineer - Transjakarta**
