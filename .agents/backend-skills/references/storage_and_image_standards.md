# S3 Storage & Image Management Standards

This document specifies the mandatory rules for handling image and file paths across all layers in CENT ACCESS (`DB`, `Service`, `REST API`, `MQTT`, `S3 Cleanup`).

---

## 1. Database Storage Standard (Pure Relative Path)

- **Rule**: All database columns storing image/file paths (`face_image_path`, `capture_image_path`, `logo_image_path`, `file_path`, etc.) MUST strictly store **Pure Relative Path** without any domain, protocol, or `S3_ROOT_PATH` (e.g. `dev/` or `cloud/`).
- **Correct in DB**:
  ```text
  tenants/123e4567-e89b-12d3-a456-426614174000/images/faces/people/avatar.jpg
  ```
- **Forbidden in DB**:
  - ❌ `https://storage.centaccess.com/dev/tenants/...` (Do not store domain)
  - ❌ `dev/tenants/...` or `cloud/tenants/...` (Do not store `S3_ROOT_PATH`)

---

## 2. Outbound Presentation Standard (Full Public URL)

- **REST API Responses (DTOs)**:
  Before returning data to the client, the Service layer MUST convert relative paths to Full Public URLs using `util.GetFullURL(...)` inside `ConvertToResponse`:
  ```go
  if person.FaceImagePath != nil && *person.FaceImagePath != "" {
      fullURL := util.GetFullURL(*person.FaceImagePath)
      response.FaceImagePath = &fullURL
  }
  ```
- **MQTT Outbound to Gateways/Devices**:
  Payloads sent to Edge Devices/Gateways (such as `/people/update`) MUST also use `util.GetFullURL(...)` so devices can download images directly via `HTTP GET`:
  ```go
  if person.FaceImagePath != nil && *person.FaceImagePath != "" {
      personData.FaceImagePath = util.GetFullURL(*person.FaceImagePath)
  }
  ```

---

## 3. Inbound Ingestion Standard (`GetRelativePath`)

- **HTTP / MQTT Input Handlers**:
  When receiving image paths from frontend forms or gateway payloads (e.g., access scan events, LPR detection), always wrap incoming paths with `util.GetRelativePath(...)` to strip away domain, query parameters, and `S3_ROOT_PATH`:
  ```go
  if req.CaptureImagePath != "" {
      cleanPath := util.GetRelativePath(req.CaptureImagePath)
      requestSend.CaptureImagePath = &cleanPath
  }
  ```

---

## 4. S3 File Upload Standard

- **Upload Implementation**:
  When files are uploaded via `util.UploadFileFromMultipart`, `util.UploadFileFromBytes`, or `util.UploadFileFromBase64`:
  1. The S3 Storage layer constructs the physical key with `S3_ROOT_PATH` (e.g., `dev/tenants/xxx/photo.jpg`) and puts the object into S3.
  2. The function returns **only the relative path** (e.g., `tenants/xxx/photo.jpg`) for saving into the Database.

---

## 5. S3 File Cleanup & Safety Standards (System Microservice)

The central file cleanup engine in `/backend/system` (`CleanUnusedFiles`) MUST adhere to 4 safety layers:

1. **Grace Period (24 Hours)**:
   Never delete files uploaded within the last 24 hours (`now.Sub(obj.LastModified) < 24h`). This prevents race conditions with users currently uploading files or filling out forms.
2. **Strict RootPath Scope**:
   Only purge objects that are strictly prefixed with the configured `S3_ROOT_PATH` (e.g., `dev/`).
3. **Safety Abort Trigger**:
   If any microservice active path provider fails to respond, or the combined active file count is 0, the cleanup job MUST abort immediately to prevent data loss.
4. **Dry-Run & Preview API**:
   - `GET /api/v1/files/cleanup/preview?prefix=` provides a dry-run preview of candidate objects without deleting them.
   - `POST /api/v1/files/cleanup` accepts `{"dryRun": true, "minAgeHours": 24}`.
   - Daily cron job runs at 03:00 AM (`0 3 * * *`) with a 24-hour grace period.
