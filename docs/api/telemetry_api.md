# East Water Telemetry Integration API Specification

**Base URL**: `http://localhost:8080` (or `https://api.eastwater.example.com`)  
**Version**: `v1.0.0`  
**Last Updated**: `2026-10-22`  
**Service Type**: Stateless API Gateway / 3rd-Party Adapter Backend (No Database)

---

## 1. Authentication & Common Headers

All protected endpoints require either an API key header or Bearer token header if `API_KEY` is configured in the environment.

| Header | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `x-api-key` | `string` | Optional | API Key assigned to the client | `ew_live_abc123...` |
| `Authorization` | `string` | Optional | Bearer Token fallback | `Bearer ew_live_abc123...` |
| `Content-Type` | `string` | Yes | MIME type of request body | `application/json` |
| `Accept` | `string` | Yes | Expected response format | `application/json` |

---

## 2. Global Response Envelope

All API responses follow the standard `APIResponse` envelope format:

```json
{
  "success": true,
  "message": "Human readable status message",
  "data": {},
  "page": {}
}
```

---

## 3. Endpoints

### `POST` /api/v1/telemetry/data

Accepts telemetry data from clients or sensor stations, transforms it into 3rd-party format, transmits it to the external 3rd-party API adapter, and returns a transaction acknowledgement.

> [!NOTE]
> This service does not store records in a local database. The request is directly validated and dispatched to the upstream 3rd party via the adapter layer (`internal/adapter/thirdparty`).

#### Request Headers
| Header | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `Content-Type` | `string` | Yes | Payload type | `application/json` |
| `x-api-key` | `string` | No | Client API Key (if enabled) | `ew_live_key_999` |

#### Request Body
| Field | Type | Required | Description | Constraints / Example |
| :--- | :--- | :---: | :--- | :--- |
| `station_code` | `string` | Yes | Unique identifier of station/device | `STATION-EW-01` |
| `parameter` | `string` | Yes | Telemetry metric parameter | `WATER_PRESSURE`, `WATER_FLOW`, `PH_LEVEL` |
| `value` | `number` | Yes | Recorded numerical value | `4.25` |
| `unit` | `string` | Yes | Measurement unit | `bar`, `m3/h`, `pH` |
| `recorded_at` | `string` (ISO 8601) | Yes | Timestamp of measurement | `2026-10-22T10:30:00Z` |
| `metadata` | `object` | No | Optional key-value telemetry attributes | `{"battery": "98%", "sensor_id": "SN-0492"}` |

##### Example Request
```json
{
  "station_code": "STATION-EW-01",
  "parameter": "WATER_PRESSURE",
  "value": 4.25,
  "unit": "bar",
  "recorded_at": "2026-10-22T10:30:00Z",
  "metadata": {
    "battery": "98%",
    "sensor_model": "PTX-500"
  }
}
```

#### Success Response (`200 OK`)
```json
{
  "success": true,
  "message": "Telemetry data transmitted successfully",
  "data": {
    "transaction_id": "c1f7b884-3c82-4ef1-925a-4eb356f1dc7b",
    "status": "SUCCESS",
    "station_code": "STATION-EW-01",
    "processed_at": "2026-10-22T10:30:01.124Z",
    "reference_no": "TKT-a91df83b",
    "extra_info": {
      "acknowledged_at": "2026-10-22T10:30:01Z",
      "remote_message": "Telemetry data ingested successfully"
    }
  }
}
```

---

### `GET` /api/v1/telemetry/data

Queries and fetches telemetry records forwarded to the 3rd-party provider.

#### Query Parameters
| Parameter | Type | Required | Default | Description |
| :--- | :--- | :---: | :---: | :--- |
| `page` | `integer` | No | `1` | Page number (min: 1) |
| `limit` | `integer` | No | `10` | Records per page (max: 100) |
| `station_code` | `string` | No | - | Filter by station code |
| `parameter` | `string` | No | - | Filter by metric name |
| `start_date` | `string` | No | - | Filter records after date (`YYYY-MM-DD`) |
| `end_date` | `string` | No | - | Filter records before date (`YYYY-MM-DD`) |

#### Success Response (`200 OK`)
```json
{
  "success": true,
  "message": "Telemetry records retrieved successfully",
  "data": [
    {
      "id": "rec-001",
      "station_code": "STATION-EW-01",
      "parameter": "WATER_PRESSURE",
      "value": 4.25,
      "unit": "bar",
      "timestamp": "2026-10-22T10:20:00Z",
      "status": "NORMAL"
    }
  ],
  "page": {
    "current_page": 1,
    "limit": 10,
    "total_pages": 1,
    "total_rows": 1
  }
}
```

---

### `GET` /api/v1/telemetry/data/:id

Retrieves full details of a specific telemetry record by its ID from the 3rd-party provider.

#### Path Parameters
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `id` | `string` | Yes | Target telemetry record ID | `rec-001` |

#### Success Response (`200 OK`)
```json
{
  "success": true,
  "message": "Telemetry record details retrieved successfully",
  "data": {
    "id": "rec-001",
    "station_code": "STATION-EW-01",
    "parameter": "WATER_PRESSURE",
    "value": 4.25,
    "unit": "bar",
    "timestamp": "2026-10-22T10:20:00Z",
    "status": "NORMAL",
    "raw_payload": {
      "battery": "98%"
    }
  }
}
```

---

### `GET` /health

Probes server liveness and upstream 3rd-party API connectivity.

#### Success Response (`200 OK`)
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "ok",
    "database": "none (stateless proxy)",
    "third_party": "connected"
  }
}
```

---

## 4. Integration Examples

### cURL

```bash
# Health check
curl -X GET http://localhost:8080/health

# Transmit telemetry data
curl -X POST http://localhost:8080/api/v1/telemetry/data \
  -H "Content-Type: application/json" \
  -d '{
    "station_code": "STATION-EW-01",
    "parameter": "WATER_PRESSURE",
    "value": 4.25,
    "unit": "bar",
    "recorded_at": "2026-10-22T10:30:00Z"
  }'

# Query telemetry records
curl -X GET "http://localhost:8080/api/v1/telemetry/data?page=1&limit=10"
```
