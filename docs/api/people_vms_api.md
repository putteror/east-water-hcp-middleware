# VMS to HikCentral Pro - People Integration API Specification

**Base URL**: `http://localhost:8080` (or `https://api.yourdomain.com`)  
**Swagger UI**: `http://localhost:8080/swagger/index.html`  
**OpenAPI Spec**: `http://localhost:8080/swagger/doc.json`  
**Version**: `v1.1.0`  
**Last Updated**: `2026-10-22`  
**Service Role**: Stateless Integration Gateway (VMS Management Software <-> HikCentral Pro)

---

## 1. Authentication & Common Headers

| Header | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `x-api-key` | `string` | Optional | API Key assigned to the VMS software | `vms_client_secret_xyz` |
| `Authorization` | `string` | Optional | Bearer Token fallback | `Bearer vms_client_secret_xyz` |
| `Accept` | `string` | Yes | Expected response format | `application/json` |

---

## 2. Standard Global Response Envelope

All API responses follow the standard `APIResponse` envelope format:

```json
{
  "success": true,
  "message": "Human readable status message",
  "data": {},
  "page": {}
}
```

---

## 3. Endpoints Overview

| Method | Path | Action Type | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/people` | Read List | ดึงรายการบุคคลทั้งหมด พร้อมค้นหาและแบ่งหน้า |
| `GET` | `/api/people/:personCode` | Read Single | ดึงข้อมูลบุคคลรายคนด้วย personCode |
| `POST` | `/api/people` | **Create** (Strict) | สร้างบุคคลใหม่ (หาก personCode ซ้ำจะคืน `409 Conflict`) |
| `PUT` | `/api/people/:personCode` | **Update** (Strict) | แก้ไขข้อมูลบุคคลที่มีอยู่แล้ว (หากไม่พบจะคืน `404 Not Found`) |
| `POST` / `PUT` | `/api/people/upsert` | **Upsert** | หากมีแล้วจะอัปเดต หากยังไม่มีจะสร้างใหม่ให้อัตโนมัติ |
| `DELETE` | `/api/people` | Batch Delete | ลบรายการบุคคลจาก personCodes array ใน JSON body |
| `DELETE` | `/api/people/:personCode` | Single Delete | ลบบุคคลรายคนจาก URL parameter |

> [!NOTE]
> ระบบรองรับทั้ง Path ที่มี prefix `/api` (เช่น `/api/people`) และเรียกตรงที่ Root (เช่น `/people`) เพื่อความสะดวกของซอฟต์แวร์ VMS

---

## 4. Endpoint Details

### 1. `GET` /api/people

ดึงรายการบุคคลทั้งหมดในระบบ พร้อมรองรับการค้นหาและแบ่งหน้า

#### Query Parameters
| Parameter | Type | Required | Default | Description |
| :--- | :--- | :---: | :---: | :--- |
| `page` | `integer` | No | `1` | หน้าที่ต้องการดึง |
| `limit` | `integer` | No | `10` | จำนวนต่อหน้า |
| `personCode` / `code` | `string` | No | - | ค้นหาจากรหัสบุคคล (รองรับทั้ง `personCode` และ `code`) |
| `personName` | `string` | No | - | ค้นหาจากชื่อบุคคล |

---

### 2. `GET` /api/people/:personCode

ดึงข้อมูลบุคคลรายบุคคลโดยตรงจาก `personCode` (รวมข้อมูลใบหน้า Base64 และป้ายทะเบียนทั้งหมด)

#### Path Parameters
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `personCode` | `string` | Yes | รหัสประจำตัวบุคคลที่ต้องการดึง | `P001` |

#### Success Response (`200 OK`)
```json
{
  "success": true,
  "message": "Person retrieved successfully",
  "data": {
    "personCode": "P001",
    "personName": "Somchai Jaidee",
    "licensePlateTexts": [
      "1กก1234"
    ],
    "deviceNames": [
      "Main Entrance Door 1",
      "Building A Entrance"
    ],
    "accessLevelIds": [
      "1",
      "2"
    ],
    "startDateTime": "2026-01-01T00:00:00Z",
    "endDateTime": "2026-12-31T23:59:59Z",
    "hasFaceImage": true,
    "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64...",
    "status": "ACTIVE",
    "processedAt": "2026-10-22T13:30:00.000Z"
  }
}
```

#### Error Response (`404 Not Found`)
```json
{
  "success": false,
  "message": "Resource not found: personCode not found in HikCentral"
}
```

---

### 3. `POST` /api/people (Strict Create)

สร้างบุคคลใหม่ (หากรหัสบุคคล `personCode` มีอยู่ในระบบ HikCentral แล้ว จะตอบกลับ `409 Conflict`)

> [!TIP]
> - หากต้องการให้ระบบสร้างหรืออัปเดตให้อัตโนมัติ สามารถส่ง Query Parameter `?upsert=true` มากับ endpoint นี้ได้ หรือใช้ endpoint `/api/people/upsert` โดยตรง
> - **การกำหนดสิทธิ์การเข้า-ออก**: สามารถเลือกส่ง `deviceNames` หรือ `accessLevelIds` (ของ HikCentral) หรือส่งคู่กันได้
>   - `deviceNames`: รองรับการส่งชื่อบางส่วน (Partial / Substring matching เช่น `"Main Entrance"` หรือ `"Building A"` ระบบจะจับคู่กับประตูในระบบให้อัตโนมัติ)
>   - `accessLevelIds`: ส่ง Access Level ID ของ HikCentral (เช่น `["1", "2"]` หรือ `[1, 2]` หรือ `"1,2"`) เหมาะสำหรับกรณีที่ไม่ต้องการระบุรายประตู

#### Content-Type: `application/json`
| Field | Type | Required | Description | Example / Constraints |
| :--- | :--- | :---: | :--- | :--- |
| `personName` | `string` | Yes | ชื่อ-นามสกุล บุคคล | `Kittisak Sukjai` |
| `personCode` | `string` | Yes | รหัสประจำตัวบุคคล (Unique) | `P003` |
| `licensePlateTexts` | `array of string` | No | ป้ายทะเบียนรถยนต์ | `["1กก9999", "2ขข1234"]` |
| `deviceNames` | `array of string` | **เลือกส่งอย่างใดอย่างหนึ่ง** | ชื่อประตู/อุปกรณ์ที่อนุญาตให้ผ่าน (ส่งอย่างใดอย่างหนึ่งระหว่าง `deviceNames` หรือ `accessLevelIds`) รองรับชื่อย่อ/ไม่เต็มคำ เช่น `"Main Entrance"` | `["Main Entrance", "Building A"]` |
| `accessLevelIds` | `array of string/int` | **เลือกส่งอย่างใดอย่างหนึ่ง** | HikCentral Access Level IDs (ส่งอย่างใดอย่างหนึ่งระหว่าง `accessLevelIds` หรือ `deviceNames`) เหมาะสำหรับกรณีไม่อยากเพิ่มทีละประตู | `["1", "2"]` หรือ `[1, 2]` |
| `startDateTime` | `string` | No | วันที่และเวลาเริ่มต้นสิทธิ์ (ISO 8601) | `2026-05-01T00:00:00Z` |
| `endDateTime` | `string` | No | วันที่และเวลาสิ้นสุดสิทธิ์ (ISO 8601) | `2026-12-31T23:59:59Z` |
| `faceImage` | `string` | No | รูปถ่ายใบหน้ารูปเดียวในรูปแบบ **Base64** (รองรับทั้งแบบ raw Base64 หรือ `data:image/jpeg;base64,...`) | `/9j/4AAQSkZJRgABAQEASABIAAD...` |

#### Request Body (JSON Example)
```json
{
  "personName": "Kittisak Sukjai",
  "personCode": "P003",
  "licensePlateTexts": [
    "1กก9999"
  ],
  "deviceNames": [
    "Main Entrance",
    "Building A"
  ],
  "accessLevelIds": [
    "1",
    "2"
  ],
  "startDateTime": "2026-05-01T00:00:00Z",
  "endDateTime": "2026-12-31T23:59:59Z",
  "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."
}
```

#### Success Response (`201 Created`)
```json
{
  "success": true,
  "message": "Person created successfully in HikCentral",
  "data": {
    "action": "created",
    "personCode": "P003",
    "personName": "Kittisak Sukjai",
    "licensePlateTexts": [
      "1กก9999"
    ],
    "deviceNames": [
      "Main Entrance Door 1",
      "Main Entrance Turnstile",
      "Building A Entrance",
      "Building A Turnstile"
    ],
    "accessLevelIds": [
      "1",
      "2"
    ],
    "startDateTime": "2026-05-01T00:00:00Z",
    "endDateTime": "2026-12-31T23:59:59Z",
    "hasFaceImage": true,
    "faceImage": "/9j/4AAQSkZJRg...",
    "status": "ACTIVE",
    "processedAt": "2026-10-22T10:45:00.123Z"
  }
}
```

---

### 4. `PUT` /api/people/:personCode (Strict Update)

อัปเดตข้อมูลของบุคคลที่มีอยู่เดิม (หากไม่พบรหัส `personCode` จะตอบกลับ `404 Not Found`)

#### Path Parameters
| Parameter | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `personCode` | `string` | Yes | รหัสบุคคลที่ต้องการอัปเดต | `P001` |

#### Content-Type: `application/json`
| Field | Type | Required | Description | Example |
| :--- | :--- | :---: | :--- | :--- |
| `personName` | `string` | No | ชื่อ-นามสกุลใหม่ (หากต้องการเปลี่ยน) | `Somchai Jaidee Updated` |
| `licensePlateTexts` | `array of string` | No | ป้ายทะเบียนรถยนต์ชุดใหม่ | `["9กก9999"]` |
| `deviceNames` | `array of string` | **เลือกส่งอย่างใดอย่างหนึ่ง** | ชื่ออุปกรณ์/ประตูชุดใหม่ (ส่งอย่างใดอย่างหนึ่งระหว่าง `deviceNames` หรือ `accessLevelIds`) รองรับ Partial Name | `["Pump Station"]` |
| `accessLevelIds` | `array of string/int` | **เลือกส่งอย่างใดอย่างหนึ่ง** | HikCentral Access Level IDs ชุดใหม่ (ส่งอย่างใดอย่างหนึ่งระหว่าง `accessLevelIds` หรือ `deviceNames`) | `["1", "3"]` |
| `startDateTime` | `string` | No | วันที่และเวลาเริ่มต้นสิทธิ์ใหม่ | `2026-01-01T00:00:00Z` |
| `endDateTime` | `string` | No | วันที่และเวลาสิ้นสุดสิทธิ์ใหม่ | `2026-12-31T23:59:59Z` |
| `faceImage` | `string` | No | รูปถ่ายใบหน้าใหม่ในรูปแบบ Base64 | `/9j/4AAQSkZJRg...` |

#### Request Body (JSON Example)
```json
{
  "personName": "Somchai Jaidee Updated",
  "licensePlateTexts": [
    "9กก9999"
  ],
  "deviceNames": [
    "Pump Station"
  ],
  "accessLevelIds": [
    "1",
    "3"
  ],
  "startDateTime": "2026-01-01T00:00:00Z",
  "endDateTime": "2026-12-31T23:59:59Z",
  "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."
}
```

#### Success Response (`200 OK`)
```json
{
  "success": true,
  "message": "Person updated successfully in HikCentral",
  "data": {
    "action": "updated",
    "personCode": "P001",
    "personName": "Somchai Jaidee Updated",
    "licensePlateTexts": [
      "9กก9999"
    ],
    "deviceNames": [
      "Pump Station Door 1"
    ],
    "accessLevelIds": [
      "1",
      "3"
    ],
    "startDateTime": "2026-01-01T00:00:00Z",
    "endDateTime": "2026-12-31T23:59:59Z",
    "hasFaceImage": true,
    "faceImage": "/9j/4AAQSkZJRg...",
    "status": "ACTIVE",
    "processedAt": "2026-10-22T10:48:00.123Z"
  }
}
```

---

### 5. `POST` หรือ `PUT` /api/people/upsert (Upsert)

สร้างบุคคลหากยังไม่มีในระบบ หรืออัปเดตข้อมูลหากมีรหัสบุคคลนั้นอยู่แล้ว

#### Content-Type: `application/json`
ฟิลด์เหมือนกับ `POST /api/people` ทุกประการ

#### Request Body (JSON Example)
```json
{
  "personCode": "P002",
  "personName": "Wichai Raksa",
  "licensePlateTexts": [
    "2ขข5678"
  ],
  "deviceNames": [
    "Warehouse"
  ],
  "accessLevelIds": [
    "2"
  ],
  "startDateTime": "2026-01-01T00:00:00Z",
  "endDateTime": "2026-12-31T23:59:59Z",
  "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."
}
```

#### Success Response:
- กรณีถูกสร้างใหม่: คืน HTTP `201 Created` พร้อม `"action": "created"`
- กรณีอัปเดตของเดิม: คืน HTTP `200 OK` พร้อม `"action": "updated"`

```json
{
  "success": true,
  "message": "Person updated successfully in HikCentral (Upsert)",
  "data": {
    "action": "updated",
    "personCode": "P002",
    "personName": "Wichai Raksa",
    "licensePlateTexts": [
      "2ขข5678"
    ],
    "deviceNames": [
      "Warehouse Security Door"
    ],
    "accessLevelIds": [
      "2"
    ],
    "hasFaceImage": true,
    "faceImage": "/9j/4AAQSkZJRg...",
    "status": "ACTIVE",
    "processedAt": "2026-10-22T10:50:00.123Z"
  }
}
```

---

### 6. `DELETE` /api/people (Batch Delete)

ลบรายการบุคคลโดยส่งอาร์เรย์ `personCodes`

#### Request Body (JSON)
```json
{
  "personCodes": [
    "P001",
    "P002"
  ]
}
```

---

## 5. cURL Integration Examples

### 1. ดึงข้อมูลบุคคลรายคนด้วย personCode (Get by code)
```bash
# วิธีที่ 1: ผ่าน URL Path parameter
curl -X GET "http://localhost:8080/api/people/P001" \
  -H "Accept: application/json"

# วิธีที่ 2: ผ่าน Query parameter
curl -X GET "http://localhost:8080/api/people?code=P001" \
  -H "Accept: application/json"
```

### 2. ดึงรายการบุคคลทั้งหมด (Get all people with pagination)
```bash
curl -X GET "http://localhost:8080/api/people?page=1&limit=10" \
  -H "Accept: application/json"
```

### 3. Create (Strict - JSON Payload with Base64 Face Image & Device/Access Levels)
```bash
curl -X POST "http://localhost:8080/api/people" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "personName": "Kittisak Sukjai",
    "personCode": "P003",
    "licensePlateTexts": ["1กก9999"],
    "deviceNames": ["Main Entrance", "Building A"],
    "accessLevelIds": ["1", "2"],
    "startDateTime": "2026-05-01T00:00:00Z",
    "endDateTime": "2026-12-31T23:59:59Z",
    "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."
  }'
```

### 4. Update (Strict - JSON Payload with AccessLevelIDs)
```bash
curl -X PUT "http://localhost:8080/api/people/P001" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "personName": "Somchai Updated",
    "licensePlateTexts": ["9กก9999"],
    "accessLevelIds": ["1", "3"],
    "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."
  }'
```

### 5. Upsert (สร้างถ้าไม่มี / อัปเดตถ้ามี)
```bash
# วิธีที่ 1: ผ่าน endpoint /upsert (รองรับทั้ง POST และ PUT)
curl -X POST "http://localhost:8080/api/people/upsert" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "personCode": "P002",
    "personName": "Wichai Raksa",
    "licensePlateTexts": ["2ขข5678"],
    "deviceNames": ["Warehouse"],
    "accessLevelIds": ["2"],
    "faceImage": "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."
  }'

# วิธีที่ 2: ผ่าน query param ?upsert=true
curl -X POST "http://localhost:8080/api/people?upsert=true" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "personCode": "P002",
    "personName": "Wichai Raksa",
    "accessLevelIds": ["1"]
  }'
```

> [!NOTE]
> ระบบยังคงรองรับการส่งแบบ `multipart/form-data` และอัปโหลดไฟล์รูปภาพผ่านฟิลด์ `faceImage` เผื่อซอฟต์แวร์ VMS ดั้งเดิมต้องการส่งเป็นไฟล์รูปภาพโดยตรง (Backwards Compatibility)

### 6. Delete (ลบตาม personCodes)
```bash
# ลบทีละคนผ่าน URL Path parameter
curl -X DELETE "http://localhost:8080/api/people/P001"

# ลบแบบกลุ่ม (Batch) ผ่าน JSON
curl -X DELETE "http://localhost:8080/api/people" \
  -H "Content-Type: application/json" \
  -d '{"personCodes": ["P001", "P002"]}'
```
