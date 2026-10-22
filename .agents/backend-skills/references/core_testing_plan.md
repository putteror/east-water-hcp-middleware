# Backend Testing Architecture & Strategy Guidelines

Comprehensive testing architecture guidelines and roadmap for Microservices within the CENT ACCESS platform.

---

## 💡 Pragmatic Unit Testing Policy & Conventions

1. **Logic-Driven Unit Testing**:
   - **Mandatory Unit Tests**: Required for any new or modified functions that contain **Business Logic, Complex Calculations, Validation Rules, or Conditional Branching (`if-else` / `switch-case`)**.
   - **Critical Function Examples**: `ProcessScan`, `CalculateParkingFee`, `ProcessAutomationTrigger`, `Authenticate`.
2. **Exemption for Simple Pass-through CRUD**:
   - Methods that simply fetch or persist data directly via Repositories without additional business logic or calculations do NOT require mandatory unit tests to prevent test maintenance bloat.
3. **Colocated Test File Placement (Idiomatic Go Layout)**:
   - All unit test files (`*_test.go`) MUST be placed directly side-by-side with their target source files inside `internal/core/service/` or `internal/interface/http/handler/`. Do NOT create nested sub-directories like `service/test/`.
4. **Categorized Sub-tests Structure (`t.Run`)**:
   - Structure unit tests using top-level `Test<Service>_<Method>` parent functions with nested `t.Run("Status - Scenario Description", func(t *testing.T) { ... })` sub-tests for clear categorization and indented terminal output.
5. **Log Suppression During Test Execution (`main_test.go`)**:
   - Include a package-level `main_test.go` containing `TestMain(m *testing.M)` with `log.SetOutput(io.Discard)` to suppress runtime `log.Printf` output during test execution while keeping production logs intact.

---

## 🏗️ Directory & File Layout Standard

Standardized test file placement across all backend microservices:

```text
├── internal/
│   ├── core/                           # Core Business Domain Layer
│   │   ├── repository/                 # Data Access Layer
│   │   │   └── mocks/                  # Centralized Repository Mocks (e.g. person_repository_mock.go)
│   │   └── service/                    # Business Logic Layer
│   │       ├── <service>.go            # Domain service logic implementation
│   │       ├── <service>_test.go       # Service Unit Tests (e.g. authentication_service_test.go)
│   │       └── main_test.go            # Package-level TestMain log suppression setup
│   ├── adapter/                        # Secondary Ports & Adapters Layer
│   │   └── mocks/                      # Centralized Adapter Mocks (e.g. event_broker_mock.go)
│   └── interface/                      # Primary Delivery / Interface Layer
│       └── http/                       # REST API Delivery Layer
│           └── handler/                # HTTP Controllers
│               ├── <entity>_handler.go # API Handler implementation
│               └── <entity>_handler_test.go # API & HTTP Handler Tests (e.g. person_handler_test.go)
├── test/                               # Testing Infrastructure Directory
│   ├── integration/                    # Database Integration Tests (e.g. process_access_control_test.go)
│   └── testutil/                       # Shared Test Utilities & Helpers
│       ├── db.go                       # Test DB Connection & Migration Lifecycles
│       ├── fixtures.go                 # Entity Seeding Helpers & Mock JWT Generation
│       └── http.go                     # Gin Router HTTP Performance Helper (PerformRequest)
└── test/scripts/                       # End-to-End & Hardware Simulators (e.g. mqtt/face_scan, email/resend)
```

---

## 📌 The 5 Testing Layers

1. **Unit Test (Service Logic Layer)**
   * **Sample Files**: `authentication_service_test.go`, `process_access_scan_service_test.go`
   * **Tooling**: `testing`, `testify/assert`, `testify/mock`, Centralized Mocks (`repository/mocks`, `adapter/mocks`)
   * **Scope**: Test isolated Business Rules, State Logic, and Validation via mocks using `t.Run` sub-tests.

2. **Integration Test (DB & GORM Query Layer)**
   * **Sample Files**: `process_access_control_integration_test.go` in `test/integration/`
   * **Tooling**: `testutil.SetupTestDB()`, PostgreSQL Test DB, `defer cleanup()`
   * **Scope**: Verify SQL Transactions, Cascading Deletes, and Foreign Key Constraints against live DB.

3. **API & Contract Test (HTTP & Router Layer)**
   * **Sample Files**: `person_handler_test.go` in `internal/interface/http/handler/`
   * **Tooling**: `net/http/httptest`, `testutil.PerformRequest`, Mock JWT Claims
   * **Scope**: Verify HTTP Status Codes, Route Middlewares, and JSON Response Envelopes (`common.SuccessResponse`).

4. **Inter-Service Communication Test (Cross-Microservice Layer)**
   * **Sample Files**: Microservice Contract / Integration Tests
   * **Tooling**: Redis Streams / PubSub Mocks, S3 Active Paths Internal API
   * **Scope**: Test Redis Pub/Sub Event broadcasting and inter-service internal HTTP API calls.

5. **Hardware & E2E Simulator Test (Real Device Simulation)**
   * **Sample Files**: CLI Scripts in `test/scripts/mqtt/face_scan/` and `test/scripts/email/resend/`
   * **Tooling**: MQTT Paho Client, Resend/SendGrid SDK
   * **Scope**: Simulate hardware payloads (LPR camera, Face Scan Terminal, Email Gateway) from outside the server.

---

## 🛠️ Shared Test Utilities (`testutil`)

* **`db.go`**: Manage Test DB connection lifecycle and AutoMigrate schemas.
* **`fixtures.go`**: Seed Entity helpers (`SeedTenant()`, `SeedPerson()`, `SeedAccessControlDevice()`) and mock JWT token generation (`GenerateTestJWT()`).
* **`http.go`**: HTTP Request execution helper for Gin Router testing (`PerformRequest()`).

---

## 🚀 Execution Commands

```bash
# Run Unit Tests in service package
go test ./internal/core/service/... -v

# Run Integration Tests
go test ./test/integration/... -v

# Run with Code Coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
