# Testing Guide - Transjakarta Fleet Management System

Panduan lengkap untuk menguji semua komponen sistem.

## Prerequisites

Pastikan semua services sudah berjalan:
```bash
docker-compose up -d
docker-compose ps
```

Semua container harus dalam status "Up".

---

## 1. Testing MQTT Publisher

### Verifikasi Publisher Mengirim Data

**Terminal 1 - Subscribe ke MQTT:**
```bash
docker exec -it transjakarta-mosquitto mosquitto_sub -t "/fleet/vehicle/+/location" -v
```

**Expected Output:**
```
/fleet/vehicle/B1234XYZ/location {"vehicle_id":"B1234XYZ","latitude":-6.2088,"longitude":106.8456,"timestamp":1715003456}
/fleet/vehicle/B5678ABC/location {"vehicle_id":"B5678ABC","latitude":-6.1751,"longitude":106.8270,"timestamp":1715003456}
/fleet/vehicle/B9012DEF/location {"vehicle_id":"B9012DEF","latitude":-6.2297,"longitude":106.8186,"timestamp":1715003458}
```

Anda harus melihat pesan baru setiap 2 detik untuk 3 kendaraan.

### Test Specific Vehicle Topic

```bash
# Subscribe hanya ke satu kendaraan
docker exec -it transjakarta-mosquitto mosquitto_sub -t "/fleet/vehicle/B1234XYZ/location" -v
```

---

## 2. Testing PostgreSQL Database

### Connect ke Database

```bash
docker exec -it transjakarta-postgres psql -U postgres -d transjakarta_fleet
```

### Test Queries

**1. Cek total records:**
```sql
SELECT COUNT(*) FROM vehicle_locations;
```

**Expected:** Angka yang terus bertambah (karena data masuk setiap 2 detik).

**2. Cek lokasi terakhir per kendaraan:**
```sql
SELECT 
    vehicle_id,
    latitude,
    longitude,
    timestamp,
    to_timestamp(timestamp) as readable_time
FROM vehicle_locations
WHERE vehicle_id IN ('B1234XYZ', 'B5678ABC', 'B9012DEF')
ORDER BY timestamp DESC
LIMIT 10;
```

**3. Cek records per kendaraan:**
```sql
SELECT 
    vehicle_id, 
    COUNT(*) as total_records,
    MIN(timestamp) as first_seen,
    MAX(timestamp) as last_seen
FROM vehicle_locations
GROUP BY vehicle_id;
```

**Expected:** Semua 3 kendaraan memiliki jumlah records yang hampir sama.

**4. Cek geofence entries (kendaraan B5678ABC di Monas):**
```sql
SELECT 
    vehicle_id,
    latitude,
    longitude,
    timestamp,
    created_at
FROM vehicle_locations
WHERE vehicle_id = 'B5678ABC'
ORDER BY timestamp DESC
LIMIT 5;
```

**Expected:** Koordinat dekat -6.1751, 106.8270 (Monas).

Exit: `\q`

---

## 3. Testing REST API

### Method 1: Using cURL

**Health Check:**
```bash
curl http://localhost:8080/health | jq
```

**Expected:**
```json
{
  "status": "healthy",
  "service": "transjakarta-fleet-management"
}
```

**Get Last Location:**
```bash
curl http://localhost:8080/api/v1/vehicles/B1234XYZ/location | jq
```

**Expected:**
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

**Get Location History:**
```bash
# Calculate timestamps
START=$(date -d '1 hour ago' +%s)
END=$(date +%s)

curl "http://localhost:8080/api/v1/vehicles/B1234XYZ/history?start=$START&end=$END" | jq
```

**Expected:** Array of location objects.

### Method 2: Using Swagger UI

1. Open browser: `http://localhost:8080/swagger/index.html`
2. Expand `/vehicles/{vehicle_id}/location` endpoint
3. Click "Try it out"
4. Enter vehicle_id: `B1234XYZ`
5. Click "Execute"
6. Verify Response Code: 200
7. Verify Response Body contains location data

### Method 3: Using Postman

1. Import collection: `Transjakarta_Fleet_API.postman_collection.json`
2. Run "Health Check" request
3. Run "Get Last Location - B1234XYZ"
4. Run "Get Location History - Last Hour"

**Expected Results:**
- All requests return 200 OK
- Response time < 1000ms
- Valid JSON responses

---

## 4. Testing RabbitMQ Geofence Events

### Via RabbitMQ Management UI

1. Open: `http://localhost:15672`
2. Login: `guest / guest`
3. Go to "Queues" tab
4. Click on `geofence_alerts` queue
5. Scroll to "Get messages" section
6. Set "Messages": 10
7. Click "Get Message(s)"

**Expected:**
- Messages with `geofence_entry` events
- JSON payload dengan:
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

### Via Command Line

**Check Queue Stats:**
```bash
docker exec transjakarta-rabbitmq rabbitmqctl list_queues name messages
```

**Expected:** `geofence_alerts` queue harus memiliki messages > 0.

### Via Logs

```bash
docker logs transjakarta-backend | grep "geofence"
```

**Expected Output:**
```
Published geofence event for vehicle B5678ABC
Received geofence event: Vehicle B5678ABC entered geofence at (-6.175100, 106.827000)
```

---

## 5. Testing Error Handling

### Test Non-Existent Vehicle

```bash
curl http://localhost:8080/api/v1/vehicles/INVALID123/location
```

**Expected:**
```json
{
  "error": "no location found for vehicle INVALID123"
}
```
Status Code: 404

### Test Missing Query Parameters

```bash
curl http://localhost:8080/api/v1/vehicles/B1234XYZ/history
```

**Expected:**
```json
{
  "error": "start and end query parameters are required"
}
```
Status Code: 400

### Test Invalid Timestamp

```bash
curl "http://localhost:8080/api/v1/vehicles/B1234XYZ/history?start=invalid&end=123456"
```

**Expected:**
```json
{
  "error": "invalid start timestamp"
}
```
Status Code: 400

---

## 6. Load Testing (Optional)

### Using Apache Bench

```bash
# Install ab (if not installed)
apt-get install apache2-utils

# Test 1000 requests with concurrency 10
ab -n 1000 -c 10 http://localhost:8080/api/v1/vehicles/B1234XYZ/location
```

**Expected:**
- No failed requests
- Average response time < 100ms
- All requests return 200 OK

### Using wrk

```bash
# Install wrk
apt-get install wrk

# Run load test for 30 seconds
wrk -t4 -c100 -d30s http://localhost:8080/api/v1/vehicles/B1234XYZ/location
```

---

## 7. Integration Testing Checklist

- [ ] MQTT Publisher sends data every 2 seconds
- [ ] Backend receives and logs MQTT messages
- [ ] Data is saved to PostgreSQL database
- [ ] GET /vehicles/{id}/location returns latest location
- [ ] GET /vehicles/{id}/history returns historical data
- [ ] Geofence detection works for vehicle B5678ABC
- [ ] RabbitMQ receives geofence events
- [ ] Worker processes geofence events from queue
- [ ] All Docker containers are running
- [ ] Health check endpoint returns 200
- [ ] Swagger documentation is accessible
- [ ] Error handling works correctly

---

## 8. Performance Benchmarks

### Expected Metrics

| Metric | Expected Value |
|--------|---------------|
| API Response Time (avg) | < 50ms |
| MQTT Message Processing | < 10ms |
| Database Insert Time | < 20ms |
| Geofence Calculation | < 5ms |
| Memory Usage (Backend) | < 100MB |
| CPU Usage (Backend) | < 10% |

### Monitor Resource Usage

```bash
# Check container stats
docker stats

# Check backend memory
docker stats transjakarta-backend --no-stream

# Check logs for errors
docker-compose logs --tail=100 | grep -i error
```

---

## 9. System Verification Script

Save this as `test.sh` and run:

```bash
#!/bin/bash

echo "=== Transjakarta Fleet Management System Test ==="
echo ""

# 1. Check Docker containers
echo "1. Checking Docker containers..."
docker-compose ps

# 2. Test Health Check
echo ""
echo "2. Testing Health Check..."
curl -s http://localhost:8080/health | jq

# 3. Test API - Last Location
echo ""
echo "3. Testing Get Last Location..."
curl -s http://localhost:8080/api/v1/vehicles/B1234XYZ/location | jq

# 4. Test API - History
echo ""
echo "4. Testing Get Location History..."
START=$(date -d '10 minutes ago' +%s)
END=$(date +%s)
curl -s "http://localhost:8080/api/v1/vehicles/B1234XYZ/history?start=$START&end=$END" | jq '. | length'

# 5. Check Database
echo ""
echo "5. Checking Database Records..."
docker exec transjakarta-postgres psql -U postgres -d transjakarta_fleet -c "SELECT vehicle_id, COUNT(*) FROM vehicle_locations GROUP BY vehicle_id;"

# 6. Check RabbitMQ Queue
echo ""
echo "6. Checking RabbitMQ Geofence Queue..."
docker exec transjakarta-rabbitmq rabbitmqctl list_queues name messages

echo ""
echo "=== Test Complete ==="
```

Make executable and run:
```bash
chmod +x test.sh
./test.sh
```

---

## 10. Troubleshooting

### Issue: No MQTT messages received

**Solution:**
```bash
# Check publisher logs
docker logs transjakarta-publisher

# Restart publisher
docker-compose restart publisher

# Manually test MQTT
docker exec -it transjakarta-mosquitto mosquitto_pub -t "/fleet/vehicle/TEST/location" -m '{"vehicle_id":"TEST","latitude":-6.2,"longitude":106.8,"timestamp":1715003456}'
```

### Issue: Database connection error

**Solution:**
```bash
# Check PostgreSQL logs
docker logs transjakarta-postgres

# Restart database
docker-compose restart postgres

# Wait for health check
docker-compose ps postgres
```

### Issue: API returns 500 error

**Solution:**
```bash
# Check backend logs
docker logs transjakarta-backend --tail 50

# Restart backend
docker-compose restart backend
```

### Issue: No geofence events in RabbitMQ

**Solution:**
1. Check if vehicle B5678ABC is near Monas coordinates
2. Check backend logs for geofence detection
3. Verify RabbitMQ connection in backend logs
4. Check geofence configuration in .env file

```bash
# Verify geofence config
cat .env | grep GEOFENCE

# Check if geofence is triggered
docker logs transjakarta-backend | grep -i "inside geofence\|published geofence"
```

---

## Summary

Untuk testing lengkap, jalankan test dalam urutan berikut:

1. ✅ Start all services: `docker-compose up -d`
2. ✅ Verify containers: `docker-compose ps`
3. ✅ Test MQTT: Subscribe dan lihat messages
4. ✅ Test Database: Connect dan query data
5. ✅ Test API: Health, Last Location, History
6. ✅ Test RabbitMQ: Check geofence queue
7. ✅ Test Error Handling: Invalid inputs
8. ✅ Monitor Logs: Check for errors
9. ✅ Run automated test script

**Total Testing Time: ~15 minutes**
