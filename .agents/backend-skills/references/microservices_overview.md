# CENT ACCESS - Microservice Architecture & Domain Map

This reference documents the responsibilities, cross-service event flows, data ownership rules, and domain gotchas for each backend service in `/backend`.

## 🔑 Cross-Service Redis Payload Standard (`refId`)

All Pub/Sub and Sync event payloads published across services MUST contain an explicit top-level `"refId"` string (representing the primary entity UUID, e.g. `registrant.ID`, `visitor.ID`, `person.ID`, `visit.ID`).

```json
{
    "refId": "18c3998c-1be6-48b1-b755-960bf2657c6d",
    "userData": { ... },
    "data": { ... }
}
```

This guarantees that `broker_redis` extracts `"ref:<uuid>"` directly for Redis stream debouncing and deduplication, eliminating random recursive ID scans over `userData` metadata.

---

## 🏛️ Microservice Domains & Responsibilities

### 1. Visitor Management Service (`/backend/visitor`)
- **Responsibility**: Manages Visitor master profiles, Visits (appointments), Visit Policies, Visit Purposes, Contact Persons, Visitor Access Cards, and Access Scans.
- **ID Alignment Rule**: `visitor.ID` IS the single authoritative UUID for a visitor. When synced with Core Person, `visitor.ID` == `corePerson.ID`. There is no separate `person_id` column or property in `visitor`. Use `id` in DTOs.
- **Freeform Fallback Rule**: Master tables for `ContactPersons` and `VisitPurposes` must not get bloated with one-off custom names. If a name is not found in master DB, store it directly in mapping tables (`contact_person_name` / `visit_purpose_name`). Omit zero UUIDs (`id,omitempty`) in DTO responses when using freeform fallback.
- **Deduplication Rule**: When resolving `ContactPersons` and `VisitPurposes` in `mapRequestToModel`, deduplicate by resolved ID or name before appending to mapping slices.
- **Events Published**: Publishes on `broadcast:visit` channel with payload types: `VISIT_CREATED`, `VISIT_UPDATED`, `VISIT_DELETED`, `VISIT_STATUS_CHANGED`. `Upsert` must publish `VISIT_CREATED` or `VISIT_UPDATED` asynchronously after transaction completion.

---

### 2. Registration Service (`/backend/register`)
- **Responsibility**: Manages public/tenant registration forms (`register_forms`) and registrant submissions (`registrants` & `registrant_answers`).
- **Approval Sync**: On registrant approval (`SyncApprovedRegistrant`), converts registrant data and syncs to `visitor` service (or `core` service).
- **Registrant Answers Rule**: `registrant.RegistrantAnswers` already contains all submitted answers (including default answers for `HIDDEN` and `FIX` fields). Adapters must NOT loop over `form.Fields` to re-append answers, avoiding duplicate answer keys.

---

### 3. Automation Service (`/backend/automation`)
- **Responsibility**: Processes automation rules triggered by Redis broadcast events and executes actions (LINE notifications, Email notifications, Print slips).
- **Triggers Supported**:
  - `VISIT_CREATED`, `VISIT_UPDATED`, `VISIT_DELETED`, `VISIT_STATUS_CHANGED` (Namespace: `visit`)
  - `REGISTRANT_REGISTERED`, `REGISTRANT_APPROVED`, `REGISTRANT_REJECTED` (Namespace: `registrant`)
  - `PERSON_ACCESS` (Namespace: `access`)
- **Direct Event Routing**: Redis stream handlers read `bPayload.Type` directly as `triggerType` without needing redundant mapping maps.
- **Macro Formatting Rule**: DATETIME fields (ISO8601 strings, `time.Time`) in macro templates (e.g. `{{visit.startAt}}`, `{{visit.endAt}}`) must automatically format as `DD/MM/YYYY HH:mm` in local timezone (e.g., `01/07/2026 18:10`).
- **Macro Registry**: Shared macro slices (`visitMacros`, `registrantMacros`, `accessMacros`) in `GetTriggerMacros` eliminate duplicate key definitions.

---

### 4. Core Service (`/backend/core`)
- **Responsibility**: Manages central organization tenants, users, accounts, doors, hardware controllers, and core person master profiles (`people`). Pure Master Data domain.

---

### 5. System Infrastructure Service (`/backend/system`)
- **Responsibility**: Centralized Infrastructure Utility Hub for the entire platform.
  - **S3 File Engine**: Accepts multipart file uploads (`POST /api/system/files/upload` / `POST /api/public/system/files/upload`) and executes the daily 03:00 AM unused file cleanup cron job across all microservices.
  - **Realtime WebSocket Gateway**: Subscribes to Redis Pub/Sub streams (`events:access`, `events:visit`, `events:registrant`, `broadcast:*`) and broadcasts live JSON payloads to external UIs, Guard Displays, and Digital Signage (`/ws/events`).
  - **System Audit Logs**: Asynchronously records system activity and audit trails into `system_logs` / `audit_logs` DB tables.

---

## 📡 Cross-Service Redis Event Map

| Producer Service | Redis Stream / Channel | Event Type (`bPayload.Type`) | Consumer Service | Purpose |
| :--- | :--- | :--- | :--- | :--- |
| `visitor` | `broadcast:visit` | `VISIT_CREATED` | `automation`, `system` | Trigger Visit rules & Broadcast WS / Log |
| `visitor` | `broadcast:visit` | `VISIT_UPDATED` | `automation`, `system` | Trigger Visit rules & Broadcast WS / Log |
| `visitor` | `broadcast:visit` | `VISIT_DELETED` | `automation`, `system` | Trigger Visit rules & Broadcast WS / Log |
| `visitor` | `broadcast:visit` | `VISIT_STATUS_CHANGED` | `automation`, `system` | Trigger Visit rules & Broadcast WS / Log |
| `register` | `broadcast:registrant` | `REGISTRANT_REGISTERED` | `automation`, `system` | Trigger Registration rules & Broadcast WS / Log |
| `register` | `broadcast:registrant` | `REGISTRANT_APPROVED` | `automation`, `system` | Trigger Registration rules & Broadcast WS / Log |
| `register` | `broadcast:registrant` | `REGISTRANT_REJECTED` | `automation`, `system` | Trigger Registration rules & Broadcast WS / Log |
| `core` | `broadcast:access` | `PERSON_ACCESS` | `automation`, `system` | Trigger Door Access rules & Broadcast WS / Log |

