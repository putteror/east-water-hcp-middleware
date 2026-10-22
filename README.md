# East Water - Stateless API Gateway & 3rd-Party Integration Backend

โครงสร้าง Backend สำหรับรับ-ส่ง API โดยเฉพาะตามหลัก **Clean Architecture (Ports & Adapters / Hexagonal Architecture)** โดย:
- **ไม่มีการเชื่อมต่อ Database** (Stateless Proxy / Integration Gateway)
- **Adapter ทำหน้าที่ติดต่อสื่อสารกับ 3rd-Party API แทน Repository**
- มี **DTO Isolation** แยก Client Request/Response ออกจาก 3rd-Party Payload อย่างสมบูรณ์

---

## 🏛️ Architecture & Data Flow

```text
┌────────────────┐       HTTP Request       ┌───────────────────────────────┐
│ Client / App   │ ───────────────────────> │  internal/interface/http      │
│ (Mobile/Web/IoT)│ <─────────────────────── │  (Handler & Middleware)       │
└────────────────┘       HTTP Response      └──────────────┬────────────────┘
                                                           │ (Strongly-typed DTO)
                                                           ▼
                                            ┌───────────────────────────────┐
                                            │  internal/core/service        │
                                            │  (Business Logic & DTO Map)   │
                                            └──────────────┬────────────────┘
                                                           │ (Port Interface)
                                                           ▼
                                            ┌───────────────────────────────┐
                                            │  internal/core/port           │
                                            │  (ThirdPartyPort Interface)   │
                                            └──────────────┬────────────────┘
                                                           │ (Implementation)
                                                           ▼
┌────────────────┐      REST / HTTP Call    ┌───────────────────────────────┐
│ External 3rd   │ <─────────────────────── │  internal/adapter/thirdparty  │
│ Party Service  │ ───────────────────────> │  (HTTP Client Adapter)        │
└────────────────┘     3rd-Party Response   └───────────────────────────────┘
```

---

## 📂 โครงสร้างไดเรกทอรี (Directory Structure)

```text
.
├── cmd/
│   └── app/
│       └── main.go                  # Entry point, DI container และ Graceful Shutdown
├── internal/
│   ├── config/                      # Centralized Configuration (os.Getenv)
│   │   └── config.go
│   ├── core/                        # Domain Core Layer (Business Rules & Ports)
│   │   ├── common/                  # API Envelopes และ Error types
│   │   │   ├── response.go          # SuccessResponse, GetDataListResponse, ErrorResponse
│   │   │   └── errors.go            # ThirdPartyError, ErrBadRequest, ErrNotFound
│   │   ├── port/                    # Outbound Port Interfaces (แทนที่ Repository)
│   │   │   ├── hikcentral_port.go   # HikCentralPort interface (Get, Create, Delete people)
│   │   │   └── third_party_port.go  # Telemetry ThirdPartyPort interface
│   │   ├── schema/                  # DTO Isolation Layer
│   │   │   ├── person_schema.go     # VMS Client DTOs (CreatePerson, DeletePeople, PersonResponse)
│   │   │   ├── hikcentral_schema.go # HikCentral Pro target payload schemas
│   │   │   ├── common_schema.go     # PaginationQuery, PaginationData, IDsRequest
│   │   │   └── client_schema.go
│   │   └── service/                 # Business Logic & DTO Mapping
│   │       ├── person_service.go    # แปลง Form-Data & Face images -> HikCentral Format
│   │       ├── person_service_test.go
│   │       └── integration_service.go
│   ├── adapter/                     # Secondary Adapters สำหรับต่อภายนอก
│   │   ├── hikcentral/              # HikCentral Pro Adapter
│   │   │   ├── adapter.go           # HTTP REST / Artemis Signature Client
│   │   │   └── mock_adapter.go      # Thread-safe In-Memory Mock Adapter
│   │   └── thirdparty/              # Telemetry HTTP Adapter
│   └── interface/                   # Primary Delivery Layer
│       └── http/
│           ├── handler/
│           │   ├── person_handler.go# Handles /api/v1/people (Get, Post multipart, Delete)
│           │   ├── person_handler_test.go
│           │   ├── integration_handler.go
│           │   └── health_handler.go
│           ├── middleware/          # CORS, Request Logger, API Key Auth
│           └── router.go            # ลงทะเบียน Routes
├── docs/
│   └── api/
│       ├── people_vms_api.md        # รายละเอียด API สำหรับ VMS Software
│       └── telemetry_api.md
├── .env.example                     # ตัวอย่าง Environment Variables
├── .env                             # ค่า Config สำหรับ Local Development
├── Dockerfile                       # Multi-stage Docker build
├── Makefile                         # คำสั่ง build, run, test
├── go.mod
└── go.sum
```

---

## 🚀 วิธีการรันโปรแกรม (Getting Started)

### 1. รันในเครื่อง (Local Run)
```bash
# รัน server (Default จะเปิด USE_MOCK_3RDPARTY=true เพื่อทดสอบได้ทันที)
go run cmd/app/main.go
```

เมื่อรันสำเร็จ Server จะพร้อมทำงานที่ `http://localhost:8080`

### 2. รัน Unit Tests
```bash
go test -v ./...
```

---

## ⚙️ การตั้งค่าเชื่อมต่อกับ 3rd-Party จริง

แก้ไขไฟล์ `.env`:
```env
# เมื่อพร้อมต่อกับระบบจริง ให้ปิด Mock
USE_MOCK_3RDPARTY=false

# ระบุ URL และ Credentials ของ 3rd-Party
THIRD_PARTY_BASE_URL=https://api.your-partner.com
THIRD_PARTY_API_KEY=your_actual_api_key
THIRD_PARTY_SECRET=your_actual_secret
THIRD_PARTY_TIMEOUT_SECONDS=30
THIRD_PARTY_DEBUG=true
```

---

## 📡 ตัวอย่างการเรียก API

### 1. Health Check
```bash
curl -X GET http://localhost:8080/health
```

### 2. ส่งข้อมูล Telemetry ไปยัง 3rd-Party
```bash
curl -X POST http://localhost:8080/api/v1/telemetry/data \
  -H "Content-Type: application/json" \
  -d '{
    "station_code": "STATION-EW-01",
    "parameter": "WATER_PRESSURE",
    "value": 4.25,
    "unit": "bar",
    "recorded_at": "2026-10-22T10:30:00Z"
  }'
```

### 3. ดึงรายการข้อมูล
```bash
curl -X GET "http://localhost:8080/api/v1/telemetry/data?page=1&limit=10"
```
