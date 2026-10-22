package port

import (
	"context"
	"eastwater-backend/internal/core/schema"
)

// ThirdPartyPort defines outbound operations to communicate with external 3rd party services.
// Core service depends on this interface, fulfilling Ports & Adapters architecture.
type ThirdPartyPort interface {
	// TransmitData sends payload to 3rd party API
	TransmitData(ctx context.Context, payload *schema.ThirdPartyIngestRequest) (*schema.ThirdPartyIngestResponse, error)

	// FetchDataList queries list of records from 3rd party API
	FetchDataList(ctx context.Context, query schema.DataSearchQuery) (*schema.ThirdPartyQueryResponse, error)

	// FetchDataByID queries single record detail from 3rd party API
	FetchDataByID(ctx context.Context, id string) (*schema.ThirdPartyItemData, error)

	// CheckHealth verifies connectivity with 3rd party API
	CheckHealth(ctx context.Context) error
}
