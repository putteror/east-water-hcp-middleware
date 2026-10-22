# Clean Architecture & Layer Rules

## 📂 Directory Context & Architecture

The system strictly adheres to **Clean Architecture (Ports & Adapters)**, dividing concerns into layers:

```text
├── cmd/
│   └── app/
│       └── main.go                     # Application Entry Point (Bootstrap, Config Loading, DI, HTTP Server Lifecycle)
├── internal/
│   ├── config/                         # Centralized Configuration Layer
│   │   └── config.go                   # Loads ALL os.Getenv into grouped structs (cfg.Database, cfg.Redis, cfg.S3, etc.)
│   ├── database/                       # Database Setup & Migrations
│   │   ├── database.go                 # GORM Connection Pool Initializer
│   │   └── migrate.go                  # Auto-migration or Golang-Migrate execution
│   ├── core/                           # Core Business Domain (STRICT RULE: Never import from interface/ or adapter/)
│   │   ├── common/                     # Global constants, API envelope wrappers, shared errors
│   │   │   ├── response.go             # SuccessResponse, GetDataListResponse, ErrorResponse
│   │   │   └── constants.go            # Enums, statuses, global constants
│   │   ├── model/                      # GORM Entities / Database Schema
│   │   │   ├── base.go                 # BaseGormModel (ID, TenantID, Timestamps, DeletedAt)
│   │   │   └── <entity>.go             # Domain entity model structs
│   │   ├── repository/                 # Data Access Ports & SQL/GORM Implementations
│   │   │   └── <entity>_repository.go  # GORM CRUD operations bound by TenantID
│   │   ├── schema/                     # DTO Isolation Layer (Requests, Responses, Search, UserData, CustomClaims)
│   │   │   ├── user_data.go            # UserData, CustomClaims struct definitions
│   │   │   ├── common_schema.go        # PaginationQuery, IDsRequest DTOs
│   │   │   └── <entity>_schema.go      # Request DTOs, Response DTOs, BasicInfo DTOs, SearchQuery DTOs
│   │   └── service/                    # Business Logic / Use Cases
│   │       ├── <entity>_service.go     # CRUD operations, Model-to-DTO conversion (ConvertToResponse)
│   │       └── file_service.go         # Microservice S3 path provider (GET /internal/active-file-paths)
│   ├── interface/                      # Primary Delivery / Interface Layer
│   │   ├── http/                       # REST API (Gin Framework)
│   │   │   ├── handler/                # HTTP Controllers (Parses requests, invokes Service, returns common envelope)
│   │   │   │   ├── <entity>_handler.go # Main entity REST handlers
│   │   │   │   └── file_handler.go     # Internal file path provider handler
│   │   │   ├── middleware/             # HTTP Middlewares
│   │   │   │   ├── jwt_auth.go         # Validates JWT, sets c.Set("userData", userData) into Gin context
│   │   │   │   └── cors.go             # CORS configuration middleware
│   │   │   └── router.go               # Gin Engine initialization & Route Registration (Public, Internal, Protected)
│   │   ├── redis/                      # Async Event Messaging (Redis Pub/Sub)
│   │   │   ├── handler/                # Redis Event Subscriber Handlers
│   │   │   └── router.go               # Redis Event Topic Subscriptions & Dispatcher
│   │   └── mqtt/                       # IoT Device Protocol Messaging
│   │       ├── handler/                # MQTT Message Handlers (e.g. Device Heartbeat, Access Events)
│   │       └── router.go               # MQTT Topic Subscriptions & Handlers Binding
│   ├── adapter/                        # Secondary Ports & External Integration Clients
│   │   ├── broker_redis.go             # BroadcastAdapter implementation (Publishing Redis Pub/Sub events)
│   │   ├── file_adapter.go             # S3 File Provider HTTP Client Adapter for System Service
│   │   └── <target_service>_adapter.go # Inter-service REST sync client adapters (e.g. CoreAdapter, VisitorAdapter)
│   └── util/                           # Helper Utilities
│       ├── authentication.go           # GetUserDataFromContext(c) context reader
│       ├── string.go                   # String manipulators, camel-case converters
│       └── upload.go                   # S3 upload utilities & path generators
├── migrations/                         # SQL Database Migration Scripts
│   ├── 000001_initial_db.up.sql        # Table DDL & Indexes creation
│   └── 000001_initial_db.down.sql      # Rollback DDL
├── Dockerfile                          # Multi-stage Docker Containerization Build Script
├── go.mod                              # Go Module Dependency Declarations
└── go.sum                              # Module Checksums
```

---

## ⚙️ Environment Variables & Configuration Rules

1. **Centralized Configuration (`internal/config`)**:
   - ALL `os.Getenv` calls MUST strictly reside inside `internal/config/config.go`.
   - Never call `os.Getenv` directly in `cmd/app/main.go`, `service`, `repository`, or `handler` layers.
2. **Required vs. Optional Configuration Strategy**:
   - **Mandatory / Required Config (Fail-Fast)**: Critical variables (`DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) MUST cause a fatal shutdown (`log.Fatalf` or `fmt.Errorf`) immediately upon service startup if missing. Never use dangerous fallback defaults for database hosts/credentials.
   - **Optional Config (Standard Fallbacks)**: Technical or convention-based variables (`PORT`, `DB_PORT`, `DB_SSL_MODE`, `REDIS_ADDRESS`, microservice URLs) SHOULD provide sensible defaults via helper functions (e.g., `getEnvWithDefault("PORT", "8080")`) to allow effortless local development without over-polluting `.env`.
3. **Database Name Handling in Microservices**:
   - Microservices MUST support service-specific DB name overrides (e.g. `VISITOR_DB_NAME`, `REGISTER_DB_NAME`) with fallback to service defaults (e.g., `cent_visitor_db`) so Docker Compose can run out-of-the-box.
4. **Single Entry Point & Grouped Config Structs (`cfg.Database`, `cfg.Redis`, `cfg.S3`)**:
   - `main.go` must only call `config.LoadConfig()` once to retrieve the consolidated application configuration.
   - Separate loader functions (e.g. `LoadS3Config()`) are forbidden in favor of grouped nested structs (`cfg.Database.GetDSN()`, `cfg.Redis.Address`, `cfg.S3`, `cfg.Services`).

5. **`main.go` Composition & Source Specification**:
   `cmd/app/main.go` serves strictly as the Orchestration and Dependency Injection Layer. All components MUST be constructed by invoking their dedicated packages:
   - **Environment & App Config**: MUST be retrieved exclusively via `config.LoadConfig()` from `internal/config/config.go`.
   - **Database Connection & DSN**: MUST use `cfg.Database.GetDSN()` and `gorm.Open(...)` or `database.NewPostgresDB(cfg)` from `internal/database`.
   - **Database Schema Migration**: MUST be invoked via `database.AutoMigrate(db)` defined in `internal/database/migrate.go`.
   - **Repositories, Services & Handlers**: MUST be instantiated via constructors (`NewRepository`, `NewService`, `NewHandler`) defined in `internal/core/repository`, `internal/core/service`, and `internal/interface/http/handler`.
   - **HTTP Router & Endpoints**: MUST be retrieved via `router.NewRouter(...)` defined in `internal/interface/http/router.go`.
   - **Server Lifecycle & Graceful Shutdown**: MUST be managed using `http.Server` with `os.Signal` (`SIGINT`, `SIGTERM`) listeners and `srv.Shutdown(ctx)`.

---

## ⛔ Core Backend Rules

### 1. No Layer Bypassing
* **Requirement:** Handlers in the `interface` layer (HTTP, MQTT, Redis) must never query the database directly using `*gorm.DB` or bypass the service layer.
* **Flow:** Execution must always flow from the outer layers to the inner layers:
  `Interface Handler -> Core Service -> Core Repository -> Database`

### 2. Interface Handler Responsibilities (Transport Adapters)
Handlers located in `internal/interface/` (`http/handler`, `redis/handler`, `mqtt/handler`) serve strictly as delivery adapters between external communication protocols and the internal domain services.

* **Primary Responsibilities of Handlers**:
  1. **Transport Parsing & Unmarshaling**: Parse HTTP query params/JSON bodies, decode Redis Pub/Sub / Stream byte payloads, or unpack MQTT topic parameters.
  2. **Metadata & Context Extraction**: Validate auth tokens, extract `UserData` from Gin context, and retrieve routing keys (`TenantID`, `RefID`, `ID`).
  3. **Delegation with Strongly-Typed DTOs**: Forward pure Request DTOs (e.g. `schema.PersonSyncPayload`, `schema.AccessScanData`) to the Service layer without passing raw `[]byte` or HTTP request objects.
  4. **Response Formatting / Transport ACKs**: Wrap the Service's Response DTOs into standard global API envelopes (`common.SuccessResponse(c, message, data)` or `common.ErrorResponse(c, status, error)`), or acknowledge Redis Stream message IDs (`r.client.XAck`).
* **Strict Handler Boundaries**:
  - ❌ Handlers MUST NOT perform business validations, access database models, or execute DB queries.
  - ❌ Handlers MUST NOT contain domain calculations or orchestration logic (all domain decisions belong inside `internal/core/service`).

### 3. GORM Models & Database Schemas
* Represent database entities in `internal/core/model`.
* GORM models must never be returned directly in API handlers or Redis broadcasts. Always convert them to Response DTOs in the service layer.

```go
package model

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;index"`
	DeviceCode string        `gorm:"type:varchar(50);not null"`
	Name      string         `gorm:"type:varchar(100);not null"`
	Status    string         `gorm:"type:varchar(20);default:'OFFLINE'"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

---

## 🔐 JWT Authentication & `UserData` Context Management

### 1. Mandatory JWT Authentication Middleware
* **Rule:** Every microservice (`core`, `visitor`, `register`, `automation`, `system`) MUST implement and apply `JWTAuthMiddleware()` from `internal/interface/http/middleware/jwt_auth.go` onto all protected route groups in `router.go`.
* **Secret Key:** Shared token validation uses `var JwtSecret = []byte("access-control-management-secret-key")`.
* **Claims:** Tokens are parsed into `*schema.CustomClaims` (containing `AccountID`, `TenantID`, `Username`, `Role`, `RegisteredClaims`).

### 2. Context Population & `UserData` Struct
* **Rule:** Upon successful JWT verification, the middleware MUST populate authenticated user data into Gin context:
  - `c.Set("userData", userData)` (using `schema.UserData` containing `UserID`/`AccountID`, `TenantID`, `Username`, `IsSuperAdmin`, `Role`).
  - `c.Set("accountID", claims.AccountID)`
  - `c.Set("tenantID", claims.TenantID)`
  - `c.Set("username", claims.Username)`
  - `c.Set("role", claims.Role)`

### 3. Context Retrieval in Handlers
* **Rule:** Handlers MUST retrieve authenticated context via `util.GetUserDataFromContext(c)` defined in `internal/util/authentication.go`.
* **No Direct JWT Parsing in Handlers:** Handlers must never perform manual JWT decoding or bypass the middleware.
* **401 Behavior:** If `JWTAuthMiddleware` is omitted on protected endpoints or context is missing, `util.GetUserDataFromContext(c)` fails and returns `401 Unauthorized`. The frontend API interceptor (`requestAPI.ts`) catches any `401 Unauthorized` response and automatically redirects the client browser to `/login`.

---

## 🖼️ S3 File Management, Retention & Cleanup Rules

### 1. Shared S3 File Retention on Entity Deletion
* **Rule:** When deleting master records (e.g., deleting a `Person` in `core`), **DO NOT immediately delete shared S3 image files** (`util.RemoveFileFromS3`). 
* **Rationale:** Other microservices (`visitor`, `register`) or historical records may reference the same S3 image path. Immediate deletion creates broken links. File deletion MUST be handled asynchronously via the `FileCleanupService`.

### 2. Isolated Microservice `FileService` & Internal API
* **Rule:** Each microservice that stores file/image paths (`visitor`, `register`) MUST isolate path retrieval queries into a dedicated `FileService` (`file_service.go`, `file_handler.go`).
* **Isolation:** Never pollute main entity CRUD services (`visitor_service.go`, `registrant_service.go`) with cleanup queries.
* **Internal Endpoint:** Expose `GET /internal/active-file-paths` (or `/api/<service>/internal/active-file-paths`) returning active S3 file path strings.

### 3. System Infrastructure `FileService` & `FileAdapter`
* **Rule:** `/backend/system` (`system`) centralizes S3 file upload (`POST /api/uploaded-files/upload`) and S3 cleanup using `FileService` (`file_service.go`, `file_handler.go`) and `FileAdapter` (`file_adapter.go`).
* **Flow:**
  1. Fetch all S3 keys under a folder prefix via `util.ListS3Objects(prefix)`.
  2. Call registered `ActiveFilePathProvider` adapters (`core`, `visitor`, `register`) via `GET /internal/active-file-paths`.
  3. Normalize paths to S3 keys in memory and purge unreferenced (unused) S3 keys via `POST /api/system/file-cleanup` or daily 03:00 AM cron job.

### 4. Standard File Component Naming
Maintain clean, alphabetical file & struct naming across services:
- **Service**: `FileService` / `FileHandler` (`file_service.go`, `file_handler.go`)
- **Adapter**: `FileAdapter` (`file_adapter.go`)

---

## 📡 Asynchronous Messaging & Hardware Protocol Architecture

The backend architecture employs two distinct messaging standards managed via Clean Architecture Adapters (`internal/adapter/`):

1. **Redis Streams (Inter-Service Async Messaging)**:
   - **Command/Sync Naming Pattern**: `[Destination]:[Resource]:[Action]` (e.g. `visitor:visit:upsert`).
   - **Domain Event Naming Pattern**: `[Source]:[Event]` (2 parts, e.g. `broadcast:access`, `broadcast:visit`).
   - **Claim-Check Deduplication Pattern**: Redis adapter implements transparent deduplication via `SETNX` pending locks and `<topic>:latest:<idKey>` payload storage to prevent stream queue bloat.

2. **MQTT Protocol (Hardware & Device Messaging)**:
   - Managed exclusively by `core` microservice via `cent-access/#` topic tree.
   - **Device Inbound (Subscriptions)**: `cent-access/+/access_scan`, `cent-access/+/license_plate_detect`.
   - **Device Outbound (Publications)**: `cent-access/{device_sn}/person/sync`, `cent-access/{device_sn}/person/delete`.

3. **WebSocket Gateway (Client Realtime Streaming)**:
   - Centralized under `system` service (`ws://<host>/ws/events/:tenantId`, `/ws/access/:tenantId`, `/ws/visit/:tenantId`).
   - Streams real-time events to frontend monitors, kiosks, and notification toasts.

> 📖 **Messaging Specification:** For the complete list of active Redis Stream topics, MQTT sub-topics, WebSocket endpoints, deduplication details, and message contracts, refer to [Messaging Specification](file:///.agents/skills/backend-skills/references/messaging_spec.md).
