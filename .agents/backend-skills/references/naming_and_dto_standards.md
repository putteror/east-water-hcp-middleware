# Naming Standards & DTO Isolation Rules

## Data Isolation & Schema Conventions

* **Requirement:** GORM models in the `model/` layer must not be returned directly to client APIs or bypass data isolation boundaries.
* **Schema Classifications & Naming Conventions:**
  * **`*Request` DTOs:** Used for parsing and validating client inputs (e.g. `DeviceRegisterRequest`). Use binding tags (e.g., `binding:"required"`) for request validation.
  * **`*Response` DTOs:** Represents the full output data format returned to the client (e.g. `DeviceResponse`). Converts internal details into client-friendly formats.
  * **`*Info` / `*BasicInfo` DTOs:** Lightweight representations of an entity containing only essential fields (e.g. `PersonInfo`, `VisitTypeBasicInfo`). These are nested within other response schemas to avoid loading heavy relations or creating recursive references.
  * **`*SearchQuery` DTOs:** Used for mapping search, query, and pagination parameters from URL params (e.g. `DeviceSearchQuery`). Should embed `PaginationQuery` for shared page and limit fields.

---

## Service Method Naming Conventions & DTO Isolation

* **Direct DTO Returns:** Services MUST perform model-to-DTO conversion internally (calling `ConvertToResponse` inside `Create`, `UpdateByID`, `GetByID`, `GetList`) and return Response DTOs (`*schema.EntityResponse` or `[]schema.EntityBasicResponse`) directly.
* **Handlers Isolation:** Handlers must NEVER call `ConvertToResponse` or map DB models to DTOs; they simply take the DTO returned from the service and wrap it into `common.SuccessResponse` or `common.GetDataListResponse`.

### Standard Service Method Signatures

| Action | Service Method Signature | Description |
| :--- | :--- | :--- |
| **Listing** | `GetList(userData schema.UserData, searchQuery schema.EntitySearchQuery) ([]schema.EntityBasicResponse, int, error)` | Retrieves multiple records with search queries & pagination, returning DTOs. |
| **Retrieve** | `GetByID(userData schema.UserData, id uuid.UUID) (*schema.EntityResponse, error)` | Retrieves a single record by UUID, returning a DTO directly. |
| **Create** | `Create(userData schema.UserData, req *schema.EntityRequest) (*schema.EntityResponse, error)` | Saves a new record, returning a DTO directly. |
| **Full Update** | `UpdateByID(userData schema.UserData, id uuid.UUID, req *schema.EntityRequest) (*schema.EntityResponse, error)` | Updates an existing record by UUID, returning a DTO directly. |
| **Delete** | `DeleteByID(userData schema.UserData, id uuid.UUID) error` | Removes a single record by UUID. |
| **Bulk Delete** | `BulkDelete(userData schema.UserData, ids []uuid.UUID) error` | Removes multiple records by UUIDs in a single transaction. |
| **Mapping** | `ConvertToResponse(entityModel *model.Entity) (*schema.EntityResponse, error)` | Model-to-response DTO converter method used internally within the Service layer. |

---

## Global Response Envelopes

All HTTP responses must use the standard `APIResponse` envelope defined in `internal/core/common/response.go`:

```go
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Page    interface{} `json:"page,omitempty"`
}
```

- Standard lists: `common.GetDataListResponse(c, message, data, pageData)`
- Single resource: `common.SuccessResponse(c, message, data)`
- Errors: `common.ErrorResponse(c, status, message)`
