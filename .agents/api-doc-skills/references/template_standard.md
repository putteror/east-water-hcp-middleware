# Standard API Documentation Markdown Template

Use this comprehensive template when creating API documentation files across any CENT ACCESS service.

---

```markdown
# [Service Name] API Specification

**Base URL**: `https://api.centaccess.io` / `http://localhost:8080`  
**Version**: `v1.0.0`  
**Service Owner**: Backend Team  
**Last Updated**: YYYY-MM-DD  

---

## 1. Overview & Architecture

Provide a concise summary of this service/module, its role within CENT ACCESS, and key interactions (e.g., Redis Pub/Sub events, Hardware Gateways, S3 Storage).

---

## 2. Authentication & Authorization

All endpoints unless explicitly marked as **Public** require authentication.

### 2.1 JWT Authentication (Web / Mobile Admin)
Include the JWT token in the `Authorization` request header:
```http
Authorization: Bearer <your_jwt_token>
```

### 2.2 HMAC Public Signature (Machine-to-Machine / Device Push)
For public endpoints (`/api/public/*`), include the following headers:

| Header | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `x-tenant-id` | UUID | Target Tenant UUID | `3fa85f64-5717-4562-b3fc-2c963f66afa6` |
| `x-timestamp` | String | Timestamp (ISO8601 or Milliseconds) | `1727690000000` |
| `x-signature` | String | HMAC SHA256 Signature | `a1b2c3d4e5...` |

---

## 3. Global Response Structure

All responses follow the unified CENT ACCESS envelope:

### 3.1 Success Response (`200 OK` / `201 Created`)
```json
{
  "code": 200,
  "message": "Operation completed successfully",
  "data": {}
}
```

### 3.2 Error Response
```json
{
  "code": 400,
  "message": "Human-readable error description",
  "error": "Detailed validation or system error"
}
```

---

## 4. Endpoints

### 4.1 [Group / Category Name]

---

#### 🔹 [Action Name]
`[METHOD]` `/api/[path]`

**Description**: What this endpoint does.  
**Access Level**: `JWT Required` | `Public (HMAC)` | `Public (Open)`  

##### Headers
| Header | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `Authorization` | String | Yes | JWT Bearer Token | `Bearer <token>` |
| `Content-Type` | String | Yes | Content format | `application/json` |

##### Path / Query Parameters
| Parameter | Location | Type | Required | Default | Description |
| :--- | :--- | :--- | :---: | :--- | :--- |
| `page` | Query | Integer | No | `1` | Page number |
| `limit` | Query | Integer | No | `20` | Records per page |
| `keyword` | Query | String | No | - | Search keyword (name, code) |

##### Request Body
| Field | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `name` | String | Yes | Entity name | `Main Gate Terminal` |
| `is_active` | Boolean | No | Active flag | `true` |

```json
{
  "name": "Main Gate Terminal",
  "is_active": true
}
```

##### Response (`200 OK`)
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "e44d82b4-17c1-4dc7-8509-5a5f66be4e37",
    "name": "Main Gate Terminal",
    "is_active": true,
    "created_at": "2026-09-30T10:00:00Z",
    "updated_at": "2026-09-30T10:00:00Z"
  }
}
```

##### Error Responses
| HTTP Status | Code | Message | Description |
| :---: | :---: | :--- | :--- |
| `400 Bad Request` | 400 | `Invalid request body` | Field validation failure |
| `401 Unauthorized` | 401 | `Unauthorized` | Invalid or expired token |
| `404 Not Found` | 404 | `Resource not found` | Target ID does not exist |

##### Example Request
```bash
curl -X POST "https://api.centaccess.io/api/v1/resource" \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Gate Terminal",
    "is_active": true
  }'
```
```
