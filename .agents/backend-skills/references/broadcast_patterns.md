# Event Broadcast Specifications & Patterns

## 📡 Redis Event Broadcast Specification

When background workers or microservices publish events over Redis Pub/Sub (e.g., `broadcast:register`, `broadcast:visit`), the payload structure MUST follow the standardized envelope structure.

### 1. Standard Broadcast Payload Structure

```json
{
  "type": "REGISTRANT_REGISTERED",
  "refId": "personId || registrantId || etc.",
  "data": {
    "id": "e4a8b1c2-3d4e-5f6a-7b8c-9d0e1f2a3b4c",
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "status": "PENDING"
  },
  "tenantId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

### 2. Key Rules for Broadcast Payloads

1. **Standard Envelope**: Root envelope contains `type` (Event Constant Name), `refId` (Optional Primary Entity Reference ID for Event Correlation/Automation), `data` (The DTO Object), and `tenantId`.
2. **Standard Correlation (`refId`)**: Use `refId` at the root level for storing the primary entity UUID (such as `personId` or `registrantId`) to enable seamless multi-trigger event correlation across Automation Services without deep object inspection.
3. **CamelCase Keys**: All keys inside the payload MUST use `camelCase`.
4. **DTO Isolation**: Pass the **Lightweight Basic Response DTO** (e.g. `schema.RegistrantBasicResponse`, `schema.VisitResponse`) into `data`, NEVER the raw GORM DB Entity Model.

---

## Broadcast Adapter Implementation Example

```go
package adapter

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/putteror/acm-register/internal/core/schema"
)

type BroadcastAdapter interface {
	PublishRegistrantRegistered(userData schema.UserData, broadcastType string, registrant *schema.RegistrantBasicResponse) error
}

type broadcastRedisAdapter struct {
	broker EventBroker
}

func NewBroadcastRedisAdapter(broker EventBroker) BroadcastAdapter {
	return &broadcastRedisAdapter{broker: broker}
}

func (a *broadcastRedisAdapter) PublishRegistrantRegistered(userData schema.UserData, broadcastType string, registrant *schema.RegistrantBasicResponse) error {
	if a.broker == nil {
		return errors.New("broker not initialized")
	}

	var refID *uuid.UUID
	if registrant != nil {
		if registrant.PersonID != nil && *registrant.PersonID != "" {
			if parsed, err := uuid.Parse(*registrant.PersonID); err == nil && parsed != uuid.Nil {
				refID = &parsed
			}
		}
		if refID == nil && registrant.ID != "" {
			if parsed, err := uuid.Parse(registrant.ID); err == nil && parsed != uuid.Nil {
				refID = &parsed
			}
		}
	}

	payload := schema.BroadcastPayload{
		Type:     broadcastType,
		RefID:    refID,
		Data:     registrant,
		TenantID: userData.TenantID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return a.broker.Publish("broadcast:register", payloadBytes)
}
```
