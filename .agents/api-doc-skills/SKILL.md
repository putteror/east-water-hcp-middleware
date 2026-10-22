---
name: api-doc-skills
description: Guidelines, standard Markdown formatting templates, and specifications for generating API documentation for CENT ACCESS.
---

# CENT ACCESS - API Documentation Skills & Markdown Standards

This skill provides standardized Markdown templates, formatting rules, and guidelines for generating clean, developer-friendly API documentation for CENT ACCESS services and endpoints.

---

## 📌 Core Documentation Principles

1. **GitHub-Flavored Markdown (GFM)**: Use tables, code blocks with syntax highlighting (`bash`, `json`, `ts`, `go`), and GFM alerts (`> [!NOTE]`, `> [!IMPORTANT]`, `> [!WARNING]`).
2. **Unified Response Envelope**: Always document responses matching CENT ACCESS standard envelopes (`common.SuccessResponse` and `common.ErrorResponse`).
3. **Explicit Auth & Header Specs**: Clearly distinguish between **JWT Authenticated** (`Authorization: Bearer <token>`), **HMAC Public Header** (`x-signature`, `x-tenant-id`, `x-timestamp`), and **Open Public**.
4. **Realistic Payload Examples**: Provide fully functional, realistic JSON payloads (using valid UUIDs, ISO 8601 timestamps `YYYY-MM-DDTHH:mm:ssZ`, and actual domain fields) rather than placeholder strings like `"string"`.

---

## 📑 Standard Endpoint Documentation Structure

Every API endpoint document MUST adhere to the following sequence:

### 1. Header & Overview
- **Title**: Clear, descriptive feature name (e.g., `Create Access Record`)
- **HTTP Method & Route**: Code formatted with method badge (e.g., `POST /api/public/access-records` or `GET /api/people`)
- **Description**: Concise summary of what the endpoint does and any background trigger/sync events.
- **Access / Auth Type**: e.g., `JWT Protected`, `Public (HMAC Signature Required)`, or `Public (No Auth)`.

### 2. Request Specifications
- **Headers Table**: `Header`, `Type`, `Required`, `Description`, `Example`
- **Path / Query Parameters Table**: `Parameter`, `Type`, `Required`, `Default`, `Description`
- **Request Body**:
  - Field description table (`Field`, `Type`, `Required`, `Description`, `Validation / Rule`)
  - Full JSON request example in a code block.

### 3. Response Specifications
- **Success Response**: HTTP Status Code (`200 OK` / `201 Created`) with JSON snippet inside CENT ACCESS envelope (`code`, `message`, `data`).
- **Data Field Definitions Table**: Explicitly explain fields inside the `data` payload.
- **Error Responses**: Document common failure scenarios (`400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `422 Unprocessable Entity`, `500 Internal Server Error`).

### 4. Integration Examples
- Provide at least a **cURL** example and optionally a **JavaScript (Fetch/Axios)** or **Go** snippet.

---

## 🛠 Standard Markdown Template

When requested to produce an API Doc, use the template below:

````markdown
# [Service/Feature Name] API Specification

**Base URL**: `https://api.yourdomain.com` (or `http://localhost:8080`)  
**Version**: `v1.0.0`  
**Last Updated**: `YYYY-MM-DD`

---

## 1. Authentication & Common Headers

### Option A: JWT Authentication (Protected Endpoints)
| Header | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `Authorization` | `string` | Yes | Bearer Token: `Bearer <jwt_token>` |
| `Content-Type` | `string` | Yes | `application/json` |

### Option B: Public HMAC Header (Public Integration Endpoints)
| Header | Type | Required | Description |
| :--- | :--- | :---: | :--- |
| `x-tenant-id` | `string` (UUID) | Yes | Target Tenant UUID |
| `x-timestamp` | `string` (Unix/ISO) | Yes | Current epoch or ISO timestamp |
| `x-signature` | `string` | Yes | HMAC SHA256 signature generated with secret key |
| `Content-Type` | `string` | Yes | `application/json` |

---

## 2. Endpoints

### `POST` /api/example-endpoint

Brief 1-2 sentence description of endpoint behavior.

> [!NOTE]
> Add important operational details or background side-effects here (e.g. triggers hardware sync, sends notification).

#### Request Headers
| Header | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `Authorization` | `string` | Yes | Bearer Token | `Bearer eyJhbGci...` |
| `Content-Type` | `string` | Yes | Payload type | `application/json` |

#### Request Body
| Field | Type | Required | Description | Example / Constraints |
| :--- | :--- | :---: | :--- | :--- |
| `name` | `string` | Yes | Full name | `John Doe` (max 100 chars) |
| `phone_number` | `string` | No | Contact phone number | `0812345678` |
| `is_active` | `boolean` | No | Active status | Default `true` |

```json
{
  "name": "John Doe",
  "phone_number": "0812345678",
  "is_active": true
}
```

#### Success Response (`200 OK` / `201 Created`)
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "name": "John Doe",
    "phone_number": "0812345678",
    "is_active": true,
    "created_at": "2026-09-30T11:00:00Z",
    "updated_at": "2026-09-30T11:00:00Z"
  }
}
```

#### Error Responses

| Status Code | Code | Message | Description |
| :---: | :---: | :--- | :--- |
| `400` | 400 | `Invalid request body` | Payload syntax error or missing required fields |
| `401` | 401 | `Unauthorized` | Missing or invalid token / signature |
| `404` | 404 | `Resource not found` | Specified ID does not exist in tenant |
| `500` | 500 | `Internal server error` | Database or unexpected server error |

```json
{
  "code": 400,
  "message": "Field 'name' is required",
  "error": "Validation failed on field 'name'"
}
```

#### Code Example

```bash
curl -X POST "https://api.yourdomain.com/api/example-endpoint" \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "phone_number": "0812345678",
    "is_active": true
  }'
```
````

---

## 📚 References & Detailed Guides

- [Standard API Document Template](file:///.agents/skills/api-doc-skills/references/template_standard.md): Ready-to-copy Markdown template for full service API documentation.
- [Authentication & Signature Specifications](file:///.agents/skills/api-doc-skills/references/hmac_and_auth_headers.md): Deep-dive into JWT claims, HMAC signature calculation, and security headers in CENT ACCESS.
