# CENT ACCESS - Authentication & Security Headers Reference

This document describes the two primary authentication mechanisms supported in CENT ACCESS APIs and how to document them.

---

## 1. JWT Authentication (Bearer Token)

Used for user-facing applications (Next.js Web Portal, Mobile Apps).

### Header Format
```http
Authorization: Bearer <jwt_token>
```

### Context Extraction
The `JWTAuthMiddleware` decodes the token claims and injects the `userData` object into the context:
```go
userData := schema.UserData{
    AccountID: &claims.AccountID,
    TenantID:  &claims.TenantID,
    Role:      claims.Role,
}
```

---

## 2. Public HMAC Signature Authentication

Used for Machine-to-Machine (M2M) communication, IoT Gateways, Edge Devices, and standalone Cron Services.

### Required Headers
| Header | Type | Description |
| :--- | :--- | :--- |
| `x-signature` | String | Base64 or Hex HMAC-SHA256 hash |
| `x-tenant-id` | UUID String | Target Tenant UUID |
| `x-timestamp` | String | Current ISO8601 or millisecond timestamp |
| `x-item-id` | UUID String (Optional) | Specific item UUID if tenant-id is omitted |

### Signature Generation Rule
The string to sign is constructed as:
```text
tenantId:<x-tenant-id>;<x-timestamp>;
```
or
```text
itemId:<x-item-id>;<x-timestamp>;
```

#### Node.js / TypeScript Example
```typescript
import * as crypto from 'crypto';

function generateSignature(tenantId: string, timestamp: string, secretKey: string): string {
  const payload = `tenantId:${tenantId};${timestamp};`;
  return crypto.createHmac('sha256', secretKey).update(payload).digest('hex');
}
```

#### Go Example
```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
)

func GenerateSignature(tenantID string, timestamp string, secretKey string) string {
    payload := fmt.Sprintf("tenantId:%s;%s;", tenantID, timestamp)
    h := hmac.New(sha256.New, []byte(secretKey))
    h.Write([]byte(payload))
    return hex.EncodeToString(h.Sum(nil))
}
```
