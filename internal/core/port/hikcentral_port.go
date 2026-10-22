package port

import (
	"context"

	"eastwater-backend/internal/core/schema"
)

// HikCentralPort defines the outbound contract for communicating with HikCentral Pro.
// Follows Ports & Adapters: Core Service depends exclusively on this interface.
type HikCentralPort interface {
	// GetPeople retrieves the list of people from HikCentral Pro
	GetPeople(ctx context.Context, query schema.PersonSearchQuery) ([]schema.HikPersonItem, int, error)

	// GetPersonByCode retrieves a single person record by personCode
	GetPersonByCode(ctx context.Context, personCode string) (*schema.HikPersonItem, error)

	// CreatePerson creates a new person record (with vehicle & face photos) in HikCentral Pro
	CreatePerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, error)

	// UpdatePerson updates an existing person record in HikCentral Pro
	UpdatePerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (*schema.HikPersonItem, error)

	// UpsertPerson creates the person if they do not exist, or updates them if they already exist
	UpsertPerson(ctx context.Context, payload *schema.HikCreatePersonPayload) (item *schema.HikPersonItem, wasCreated bool, err error)

	// DeletePeople deletes multiple people by personCodes in HikCentral Pro
	DeletePeople(ctx context.Context, personCodes []string) (deletedCodes []string, failedCodes []string, err error)

	// CheckHealth verifies connectivity with HikCentral Pro server
	CheckHealth(ctx context.Context) error
}
