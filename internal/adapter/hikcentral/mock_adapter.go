package hikcentral

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/schema"
	"github.com/google/uuid"
)

var _ port.HikCentralPort = (*MockAdapter)(nil)

// MockAdapter provides thread-safe in-memory store simulating HikCentral Pro OpenAPI
type MockAdapter struct {
	mu     sync.RWMutex
	people map[string]*schema.HikPersonItem // Keyed by PersonCode
}

// NewMockAdapter initializes a mock HikCentral Pro adapter with initial sample records
func NewMockAdapter() *MockAdapter {
	m := &MockAdapter{
		people: make(map[string]*schema.HikPersonItem),
	}

	// Seed with sample people
	m.people["P001"] = &schema.HikPersonItem{
		AppointRecordID: "APPT-001",
		PersonID:        uuid.New().String(),
		PersonCode:      "P001",
		PersonName:      "Somchai Jaidee",
		BeginTime:       "2026-01-01T00:00:00Z",
		EndTime:         "2026-12-31T23:59:59Z",
		Vehicles: []schema.HikVehicleData{
			{PlateNo: "1กก1234"},
		},
		DeviceNames:    []string{"Main Entrance Door 1", "Building A Turnstile"},
		AccessLevelIDs: []string{"1"},
		HasFaceImage:   true,
		FaceBase64:     "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64...",
		Status:         "ACTIVE",
	}

	m.people["P002"] = &schema.HikPersonItem{
		AppointRecordID: "APPT-002",
		PersonID:        uuid.New().String(),
		PersonCode:      "P002",
		PersonName:      "Wichai Raksa",
		BeginTime:       "2026-02-01T00:00:00Z",
		EndTime:         "2026-11-30T23:59:59Z",
		Vehicles: []schema.HikVehicleData{
			{PlateNo: "2ขข5678"},
		},
		DeviceNames:    []string{"East Water Gate Barrier In", "Pump Station Door 1"},
		AccessLevelIDs: []string{"1", "2"},
		HasFaceImage:   true,
		FaceBase64:     "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64P2...",
		Status:         "ACTIVE",
	}

	return m
}

func (m *MockAdapter) GetPeople(ctx context.Context, query schema.PersonSearchQuery) ([]schema.HikPersonItem, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []schema.HikPersonItem
	for _, p := range m.people {
		if query.PersonCode != "" && !strings.Contains(strings.ToLower(p.PersonCode), strings.ToLower(query.PersonCode)) {
			continue
		}
		if query.PersonName != "" && !strings.Contains(strings.ToLower(p.PersonName), strings.ToLower(query.PersonName)) {
			continue
		}
		matched = append(matched, *p)
	}

	total := len(matched)

	// Apply pagination slice
	start := (query.Page - 1) * query.Limit
	if start >= total {
		return []schema.HikPersonItem{}, total, nil
	}
	end := start + query.Limit
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (m *MockAdapter) GetPersonByCode(ctx context.Context, personCode string) (*schema.HikPersonItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.people[personCode]
	if !exists {
		return nil, common.ErrNotFound
	}
	return item, nil
}

func (m *MockAdapter) CreatePerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.people[payload.PersonCode]; exists {
		return nil, fmt.Errorf("%w: personCode %s already exists in HikCentral", common.ErrAlreadyExists, payload.PersonCode)
	}

	faceStr := payload.GetFaceBase64()
	item := &schema.HikPersonItem{
		AppointRecordID: "APPT-" + uuid.New().String()[:8],
		PersonID:        uuid.New().String(),
		PersonCode:      payload.PersonCode,
		PersonName:      payload.PersonName,
		BeginTime:       payload.BeginTime,
		EndTime:         payload.EndTime,
		Vehicles:        payload.Vehicles,
		DeviceNames:     payload.DeviceNames,
		AccessLevelIDs:  payload.AccessLevelIDs,
		HasFaceImage:    faceStr != "",
		FaceBase64:      faceStr,
		Status:          "ACTIVE",
	}

	m.people[payload.PersonCode] = item
	return item, nil
}

func (m *MockAdapter) UpdatePerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.people[payload.PersonCode]
	if !exists {
		return nil, fmt.Errorf("%w: personCode %s does not exist in HikCentral", common.ErrNotFound, payload.PersonCode)
	}

	if payload.PersonName != "" {
		existing.PersonName = payload.PersonName
	}
	if payload.BeginTime != "" {
		existing.BeginTime = payload.BeginTime
	}
	if payload.EndTime != "" {
		existing.EndTime = payload.EndTime
	}
	if len(payload.Vehicles) > 0 {
		existing.Vehicles = payload.Vehicles
	}
	if len(payload.DeviceNames) > 0 {
		existing.DeviceNames = payload.DeviceNames
	}
	if len(payload.AccessLevelIDs) > 0 {
		existing.AccessLevelIDs = payload.AccessLevelIDs
	}
	if faceStr := payload.GetFaceBase64(); faceStr != "" {
		existing.FaceBase64 = faceStr
		existing.HasFaceImage = true
	}

	return existing, nil
}

func (m *MockAdapter) UpsertPerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.people[payload.PersonCode]
	if exists {
		// Update existing
		if payload.PersonName != "" {
			existing.PersonName = payload.PersonName
		}
		if payload.BeginTime != "" {
			existing.BeginTime = payload.BeginTime
		}
		if payload.EndTime != "" {
			existing.EndTime = payload.EndTime
		}
		if len(payload.Vehicles) > 0 {
			existing.Vehicles = payload.Vehicles
		}
		if len(payload.DeviceNames) > 0 {
			existing.DeviceNames = payload.DeviceNames
		}
		if len(payload.AccessLevelIDs) > 0 {
			existing.AccessLevelIDs = payload.AccessLevelIDs
		}
		if len(payload.FaceBase64s) > 0 {
			existing.FaceBase64s = payload.FaceBase64s
			existing.FacesCount = len(payload.FaceBase64s)
		}
		if faceStr := payload.GetFaceBase64(); faceStr != "" {
			existing.FaceBase64 = faceStr
			existing.HasFaceImage = true
		}
		return existing, false, nil
	}

	// Create new
	faceStr := payload.GetFaceBase64()
	item := &schema.HikPersonItem{
		AppointRecordID: "APPT-" + uuid.New().String()[:8],
		PersonID:        uuid.New().String(),
		PersonCode:      payload.PersonCode,
		PersonName:      payload.PersonName,
		BeginTime:       payload.BeginTime,
		EndTime:         payload.EndTime,
		Vehicles:        payload.Vehicles,
		DeviceNames:     payload.DeviceNames,
		AccessLevelIDs:  payload.AccessLevelIDs,
		HasFaceImage:    faceStr != "",
		FaceBase64:      faceStr,
		FacesCount:      len(payload.FaceBase64s),
		FaceBase64s:     payload.FaceBase64s,
		Status:          "ACTIVE",
	}

	m.people[payload.PersonCode] = item
	return item, true, nil
}

func (m *MockAdapter) DeletePeople(ctx context.Context, personCodes []string) ([]string, []string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var deleted []string
	var failed []string

	for _, code := range personCodes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		if _, exists := m.people[trimmed]; exists {
			delete(m.people, trimmed)
			deleted = append(deleted, trimmed)
		} else {
			failed = append(failed, trimmed)
		}
	}

	if len(deleted) == 0 && len(failed) > 0 {
		return nil, failed, fmt.Errorf("%w: none of the requested personCodes were found in HikCentral", common.ErrNotFound)
	}

	return deleted, failed, nil
}

func (m *MockAdapter) CheckHealth(ctx context.Context) error {
	return nil
}
