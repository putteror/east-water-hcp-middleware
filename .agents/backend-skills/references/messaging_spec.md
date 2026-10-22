# Backend Pub/Sub & Async Messaging Catalog

This document catalogs all messaging topics used for asynchronous communication across the CentAccess platform via **Redis Streams** (inter-service backend communication) and **MQTT** (communication between backend and hardware access control devices/gateways).

---

## 📌 1. Redis Streams (Inter-Service Communication)

Used for issuing commands to sync data or broadcasting asynchronous domain events between backend microservices.

### 🧭 Naming Conventions
* **Command & Data Sync Streams:** Named based on **Destination-first** in the format `[Destination] : [Resource] : [Action]`
* **Domain Events Streams:** Named based on **Source-first** or shared event topic in the format `[Source] : [Event]`

| Topic / Stream Name | Publisher | Subscriber | Claim-Check (Deduplication) | Description |
| :--- | :--- | :--- | :---: | :--- |
| **`visitor:visit:upsert`** | `register` | `visitor-management` | **Yes** | Transmits approved registrant data to create or update Visit records. |
| **`core:person:upsert`** | `visitor-management` | `core` | **Yes** | Transmits approved visitor profile to create/update master `Person` record. |
| **`core:person:delete`** | `visitor-management` | `core` | **Yes** | Command to delete or revoke `Person` access privileges from master DB. |
| **`core:person:temporary`** | `event`, `visitor` | `core` | **Yes** | Command to grant or revoke temporary access privileges (rules / devices with time window). |
| **`system:audit_log:create`** | Microservices | `system` | **No** | Asynchronously pushes audit log records from all services to `system` DB. |
| **`broadcast:person`** | `core` | `event`, `visitor`, `system` | **No** | Broadcasts master Person mutations (created, updated, deleted) to listening services. |
| **`broadcast:access`** | `core` | `automation`, `overtime`, `system`, `visitor` | **No** | Real-time door scan events for automation (LINE/Email) and OT calculation. |
| **`broadcast:device_status`** | `core` | `system` | **No** | Real-time hardware device status and gateway discovery events. |
| **`broadcast:register`** | `register` | `automation` | **No** | Visitor registration events for automation triggers. |
| **`broadcast:visit`** | `visitor` | `automation` | **No** | Visitor check-in/visit lifecycle events. |

---

## 🔄 Redis Streams Claim-Check Deduplication Pattern

To prevent redundant processing during rapid bursts of entity updates, the system enforces the Claim-Check / Deduplication Pattern:

1. **Publisher Side (`Publish`)**:
   - Extract unique identifier key (e.g., `person:123`).
   - Store latest JSON payload in Redis Key `<topic>:latest:<idKey>` (24h TTL).
   - Create pending lock key `<topic>:pending:<idKey>` via `SETNX`. If successful, publish a single message containing `idKey` to the Stream. If lock exists, skip adding duplicate stream messages.

2. **Subscriber Side (`Subscribe`)**:
   - Consumer Groups read stream messages via `XReadGroup`.
   - Read latest payload from `<topic>:latest:<idKey>`.
   - Delete pending lock `<topic>:pending:<idKey>`.
   - Execute handler logic and call `XACK`.

---

## 📡 2. MQTT Topics & Architecture (Hardware & Gateway Communication)

Backend Core utilizes `internal/adapter/broker_mqtt.go` (`adapter.MQTTBroker`) for outbound commands and `internal/interface/mqtt/` (`MQTTRouter`) for inbound event subscription.

### 📦 Standard MQTT Payload Envelope (`schema.MQTTPayload`)

All MQTT payloads exchanged between Backend Core and Gateways adhere to the standard JSON envelope structure with support for Async Request-Response via `responseTopic`:

```json
{
  "transactionId": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "tenantId": "c019283f-1234-5678-90ab-cdef01234567",
  "gatewayId": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "responseTopic": "optional/reply/topic",
  "data": {
    /* Feature-specific DTO payload */
  }
}
```

---

### 📥 Subscriptions (Receiving Events from Gateway ➔ Backend Core)
Core Service listens to events via wildcard topic `core/+#` and `cent-access/#` (legacy fallback):

| Sub-topic Pattern | Sender | Receiver | Description |
| :--- | :--- | :--- | :--- |
| `core/event/access-scan` | Gateway / Scanner | `core` (MQTT Router) | Transmits door access scan event (Face / Card / QR). |
| `core/event/license-plate` | Gateway / LPR | `core` (MQTT Router) | Transmits LPR camera vehicle license plate detection. |
| `core/gateway/status` | Gateway | `core` (MQTT Router) | Gateway transmits Heartbeat status, CPU/RAM usage, and device health. |
| `core/log/integration` | Gateway | `core` (MQTT Router) | Gateway transmits hardware operation logs and failure diagnostics. |
| `integration-gateway/discover/{tenantId}` | Gateway | `core` (MQTT Router) | Unpaired gateway broadcast discovery signal. |

---

### 📤 Publications (Sending Commands from Backend Core ➔ Targeted Gateway)
Core Service publishes targeted commands directly to specific Gateway UUID (`{gatewayId}`):

| Topic Name | Sender | Receiver | Description |
| :--- | :--- | :--- | :--- |
| `{gatewayId}/people/update` | `core` | Targeted Gateway | Downloads/syncs person access credentials to Gateway devices. |
| `{gatewayId}/people/remove` | `core` | Targeted Gateway | Commands Gateway to revoke user access privileges. |
| `{gatewayId}/devices/sync` | `core` | Targeted Gateway | Syncs access control device configuration & IP addresses. |
| `{gatewayId}/devices/set-callback` | `core` | Targeted Gateway | Configures door scanner callback URLs on Gateway. |
| `{gatewayId}/gateway/pair` | `core` | Targeted Gateway | Commands Gateway to pair with specified Tenant. |
| `{gatewayId}/gateway/unpair` | `core` | Targeted Gateway | Commands Gateway to unpair from Tenant. |
| `{gatewayId}/gateway/sync` | `core` | Targeted Gateway | Triggers full Gateway system re-synchronization. |
| `{gatewayId}/image/check-quality` | `core` | Targeted Gateway | Requests face image quality validation. |

---

## 🌐 3. WebSocket Gateway & Realtime Events

Used for real-time push notifications from backend to Frontend UIs (Live Monitors, Kiosks, Gate Displays), powered by **System Infrastructure Service (`system`)**.

### 🔌 WebSocket Connection Endpoints
Frontend connects to WebSocket endpoints exposed at `/backend/system`:

| Endpoint Pattern | Description | Targeted Clients |
| :--- | :--- | :--- |
| `ws://<host>/ws/events/:tenantId` | Aggregated real-time event stream for a tenant | Main Dashboard, General Notifier |
| `ws://<host>/ws/access/:tenantId` | Real-time door access scan stream | Gate Monitor, Access Kiosk |
| `ws://<host>/ws/visit/:tenantId` | Real-time visitor check-in/out stream | Reception Monitor |

### 📦 WebSocket Payload Structure
All WebSocket broadcast messages are wrapped in the standard JSON envelope:

```json
{
  "event": "access_scan | visit_checkin | registrant_approve",
  "tenantId": "uuid-string",
  "timestamp": "ISO-8601-timestamp",
  "data": {
    /* DTO Payload (e.g. AccessScanData, VisitData) */
  }
}
```
