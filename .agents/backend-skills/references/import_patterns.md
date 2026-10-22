# CentAccess - Excel & Batch Import Architecture & Patterns

This document defines the standardized pattern for implementing batch data and Excel imports across backend services in CentAccess.

---

## 1. Core Principles

1. **Separation of Concerns**:
   - For complex entities with multi-step or asynchronous batch processing (e.g. `Person`, `Device`), separate the import business logic into an isolated file (`<entity>_import.go`) instead of bloating `<entity>_service.go`.
2. **Master Data Pre-caching ($O(1)$ Lookups)**:
   - **Never query the database row-by-row** ($N+1$ anti-pattern).
   - Before iterating rows, load all existing master data into memory lookup maps (keyed by normalized lowercase names):
     ```go
     companyMap := make(map[string]uuid.UUID)
     for _, c := range companies {
         companyMap[strings.ToLower(strings.TrimSpace(c.Name))] = c.ID
     }
     ```
3. **Smart Matching & Graceful Fallback**:
   - Match fields by **Name** rather than requiring raw database UUIDs.
   - If a name matches existing master data $\rightarrow$ automatically link the `ID` and assign the name.
   - If a name is not found in master data $\rightarrow$ store the text value in the entity's free text column (e.g. `company`, `department`, `job_position`) without creating unwanted dummy records in master tables.
4. **Robust Header & Date Normalization**:
   - Clean column headers by stripping whitespace, underscores, asterisks, parentheses, and slashes:
     ```go
     header = strings.TrimSpace(strings.ToLower(headers[i]))
     header = strings.ReplaceAll(header, "_", "")
     header = strings.ReplaceAll(header, " ", "")
     header = strings.ReplaceAll(header, "*", "")
     header = strings.ReplaceAll(header, "(", "")
     header = strings.ReplaceAll(header, ")", "")
     ```
   - Parse dates using flexible layouts (`YYYY-MM-DD`, `DD/MM/YYYY`, `DD-MM-YYYY`, timestamps).
5. **Dynamic On-the-fly Template Generation**:
   - Provide an endpoint `GET /api/<entity>/import/template` that dynamically generates a styled `.xlsx` file using `excelize/v2`.
   - Include clear header styling (Dark Slate `#1E293B` with bold white text) and at least 2 realistic sample rows.
6. **Asynchronous Batch Job Tracking (`batch_jobs`)**:
   - For long-running imports (>100 records or with face image zips), execute the loop in a background goroutine and track progress via `batchJobService.RecordProgress(...)`.
