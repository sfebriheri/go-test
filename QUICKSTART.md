# Quick Start Guide - Transjakarta Fleet Management

Panduan cepat untuk menjalankan sistem dalam 5 menit.

## 🚀 Prerequisites

- Docker Desktop installed (Windows/Mac) atau Docker + Docker Compose (Linux)
- Git
- Port yang tersedia: 8080, 5432, 1883, 5672, 15672

## ⚡ Quick Start (5 Minutes)

### 1. Clone Repository

```bash
git clone <repository-url>
cd transjakarta-fleet-management
```

### 2. Start All Services

```bash
docker-compose up --build -d
```

Tunggu ~2-3 menit untuk semua services siap.

### 3. Verify Services Running

```bash
docker-compose ps
```

Semua services harus berstatus "Up".

### 4. Test the System

**Test Health Check:**
```bash
curl http://localhost:8080/health
```

**Test Get Location:**
```bash
curl http://localhost:8080/api/v1/vehicles/B1234XYZ/location | jq
```

**View Live MQTT Messages:**
```bash
docker exec -it transjakarta-mosquitto mosquitto_sub -t "/fleet/vehicle/+/location" -v
```

### 5. Access Web Interfaces

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **RabbitMQ Console**: http://localhost:15672 (guest/guest)

## 📊 What's Running?

| Service | Port | Description |
|---------|------|-------------|
| Backend API | 8080 | REST API endpoints |
| PostgreSQL | 5432 | Database |
| MQTT Broker | 1883 | Message broker |
| RabbitMQ | 5672 | Event queue |
| RabbitMQ UI | 15672 | Management console |

## 🎯 Key Features to Test

1. **Real-time Location Tracking**
   - Publisher sends data every 2 seconds
   - 3 vehicles: B1234XYZ, B5678ABC, B9012DEF

2. **REST API**
   - Get last location: `/api/v1/vehicles/{id}/location`
   - Get history: `/api/v1/vehicles/{id}/history?start={ts}&end={ts}`

3. **Geofencing**
   - Vehicle B5678ABC starts inside geofence (Monas)
   - Check RabbitMQ queue `geofence_alerts` for events

## 🛑 Stop Services

```bash
docker-compose down
```

To remove all data:
```bash
docker-compose down -v
```

## 📚 Next Steps

- Read [README.md](README.md) for detailed documentation
- Check [TESTING_GUIDE.md](TESTING_GUIDE.md) for comprehensive testing
- Follow [VIDEO_DEMO_SCRIPT.md](VIDEO_DEMO_SCRIPT.md) for demo recording
- Import Postman collection for API testing

## ⚠️ Troubleshooting

**Services not starting?**
```bash
docker-compose logs
```

**Port already in use?**
Edit `docker-compose.yml` and change port mappings.

**Database connection failed?**
```bash
docker-compose restart postgres
docker-compose logs postgres
```

## 💡 Tips

- Use `make` commands for common operations (see Makefile)
- Check logs: `docker-compose logs -f backend`
- Monitor resources: `docker stats`
- Access database: `make db-connect`

## 📞 Support

For issues, check:
1. Docker is running
2. Ports are not in use
3. Enough disk space (>2GB)
4. Logs for errors: `docker-compose logs`
