---
name: backend-skills
description: Guidelines, naming standards, DTO isolation rules, and architectural patterns for Go backend code in /backend of CENT ACCESS.
---

# CENT ACCESS - Backend Development Skills & Patterns

This skill provides comprehensive rules, standards, and architectural patterns for Go backend services within `/backend`.

## Core Guidelines Overview

1. **Clean Architecture (Ports & Adapters)**: Strict layer boundaries (`Interface -> Service -> Repository -> DB`). Direct DB queries in handlers are strictly forbidden. Outbound communication brokers MUST be isolated in `internal/adapter/` (`EventBroker` in `broker_redis.go` and `MQTTBroker` in `broker_mqtt.go`), while inbound entry points reside in `internal/interface/` (`mqtt/`, `redis/`, `http/`, `ws/`).
2. **Interface Handler Responsibilities**:
   - Handlers in `internal/interface/` (HTTP, Redis, MQTT) serve strictly as transport adapters.
   - **Their responsibilities are exclusively**: (1) parse/unmarshal transport payloads (HTTP JSON/form, Redis JSON byte streams, MQTT topics/payloads), (2) perform initial payload syntax validation & extract auth context/metadata (`UserData`, `TenantID`, `RefID`), (3) invoke the Service layer passing **strongly-typed Request DTOs**, and (4) wrap Service responses into transport-specific envelopes (e.g. `common.SuccessResponse`) or handle transport ACKs.
   - Handlers MUST NEVER contain business logic, direct DB queries, or direct inter-service orchestration.
3. **Service Method Naming & DTO Isolation**: Standardized CRUD names (`GetList`, `GetByID`, `Create`, `UpdateByID`, `DeleteByID`, `BulkDelete`). Services MUST call `ConvertToResponse` internally and return Response DTOs directly. Handlers only wrap DTOs into global API envelopes (`common.SuccessResponse`).
4. **Event Broadcasting**: Redis Pub/Sub events must use lightweight `*schema.EntityResponse` or `*schema.EntityBasicResponse` DTOs inside `data`, without field redundancy or raw GORM models.
4. **S3 File Engine, Image Standards & Safety Cleanup**: 
   - **Database Layer**: Strictly store **Pure Relative Path** without domain or `S3_ROOT_PATH` (e.g., `tenants/...`).
   - **Outbound (DTO & MQTT)**: Convert to **Full Public URL** via `util.GetFullURL(...)` before sending to REST clients or Gateways.
   - **Inbound**: Strip domain/rootPath via `util.GetRelativePath(...)` before persisting to DB.
   - **Cleanup Safety**: Daily 03:00 AM S3 Cleanup Cron in `/backend/system` enforces a **24-hour grace period**, strict rootPath scoping, and zero-count abort triggers.
5. **Environment Variables & `main.go` Composition**: ALL `os.Getenv` calls MUST strictly reside inside `internal/config/config.go`. `main.go` serves purely as an orchestrator and MUST load config via `config.LoadConfig()` (using Grouped Structs `cfg.Database.GetDSN()`, `cfg.Redis`, `cfg.S3`, `cfg.Services`), run migrations via `database.AutoMigrate(db)` from `internal/database/migrate.go`, load routes via `router.NewRouter(...)` from `internal/interface/http/router.go`, and manage lifecycle using Graceful Shutdown via `http.Server`.
6. **JWT Authentication & `UserData` Context Management**: All protected HTTP route groups in `router.go` MUST apply `protected.Use(middleware.JWTAuthMiddleware())` defined in `internal/interface/http/middleware/jwt_auth.go`. The middleware validates the `Authorization: Bearer <token>` header against `JwtSecret` using `schema.CustomClaims` and sets `c.Set("userData", userData)` into Gin context. Handlers retrieve user context via `util.GetUserDataFromContext(c)`. Missing middleware or context returns `401 Unauthorized`, triggering automatic client redirection to `/login`.
7. **Pragmatic Unit Testing Policy & Standards**: Write Unit Tests for newly created or modified functions whenever they contain non-trivial Business Logic, complex calculations, validation rules, or conditional branching (`if-else`/`switch-case`). Place `*_test.go` files side-by-side with target source files (colocated), structure test scenarios using `t.Run("Description", func(t *testing.T) { ... })` sub-tests for clear categorization, and add package `main_test.go` with `log.SetOutput(io.Discard)` inside `TestMain` to suppress test log output.
8. **API Documentation Synchronization (`docs/api/`)**: Whenever backend endpoints, Request/Response DTO schemas, query parameters, or authentication methods are added or modified, corresponding Markdown documentation under `docs/api/` MUST be updated (or created) following [.agents/skills/api-doc-skills/SKILL.md](file:///.agents/skills/api-doc-skills/SKILL.md) to keep documentation in sync with active implementations.

---

## Detailed References

For detailed guides, code templates, and implementation examples, refer to:

- [Architecture & Layer Rules](file:///.agents/skills/backend-skills/references/architecture.md): Clean Architecture, directory context, layer boundaries, JWT Auth middleware, and GORM model guidelines.
- [S3 Storage & Image Standards](file:///.agents/skills/backend-skills/references/storage_and_image_standards.md): Pure relative path DB standards, `GetFullURL` DTO conversion, MQTT outbound rules, and 24h grace period S3 safety cleanup.
- [Core Master Testing Plan](file:///.agents/skills/backend-skills/references/core_testing_plan.md): Complete testing plan, pragmatic unit testing rules, sub-test standards, and guidelines for backend testing.
- [Messaging Specification](file:///.agents/skills/backend-skills/references/messaging_spec.md): Complete specification of Redis Streams topics, Claim-Check deduplication flow, and MQTT hardware topics.
- [Microservice Architecture & Domain Map](file:///.agents/skills/backend-skills/references/microservices_overview.md): Domain responsibilities, ID alignment, Redis Pub/Sub events, and cross-service data rules.
- [Naming Standards & DTO Isolation](file:///Users/putteror/Developer/Platform/CentAccess/App/.agents/skills/backend-skills/references/naming_and_dto_standards.md): Service naming conventions, `ConvertToResponse` practices, and DTO isolation rules.
- [Excel & Batch Import Patterns](file:///Users/putteror/Developer/Platform/CentAccess/App/.agents/skills/backend-skills/references/import_patterns.md): Standardized guidelines for Pre-caching, smart name mapping, robust header parsing, and dynamic `.xlsx` template generation.
- [Event Broadcast Specifications](file:///Users/putteror/Developer/Platform/CentAccess/App/.agents/skills/backend-skills/references/broadcast_patterns.md): Standard Redis Pub/Sub payload structure, channel topics, and adapter examples.
- [Deployment & Build Reference](file:///.agents/skills/backend-skills/references/deployment.md): Fast Go cross-compilation with `build-and-push.sh` and zero-downtime selective container deployments via `docker-compose.deploy.yml`.
