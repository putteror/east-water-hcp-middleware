package schema

import "time"

// TransmitDataRequest represents incoming payload from client to be forwarded to 3rd party
type TransmitDataRequest struct {
	StationCode string                 `json:"station_code" binding:"required"`
	Parameter   string                 `json:"parameter" binding:"required"`
	Value       float64                `json:"value" binding:"required"`
	Unit        string                 `json:"unit" binding:"required"`
	RecordedAt  time.Time              `json:"recorded_at" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TransmitDataResponse represents the formatted response returned to the client
type TransmitDataResponse struct {
	TransactionID string                 `json:"transaction_id"`
	Status        string                 `json:"status"`
	StationCode   string                 `json:"station_code"`
	ProcessedAt   time.Time              `json:"processed_at"`
	ReferenceNo   string                 `json:"reference_no,omitempty"`
	ExtraInfo     map[string]interface{} `json:"extra_info,omitempty"`
}

// DataSearchQuery represents search query params received from client
type DataSearchQuery struct {
	PaginationQuery
	StationCode string `form:"station_code"`
	Parameter   string `form:"parameter"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
}

// DataBasicInfo represents summary data item in list responses
type DataBasicInfo struct {
	ID          string    `json:"id"`
	StationCode string    `json:"station_code"`
	Parameter   string    `json:"parameter"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

// DataDetailResponse represents detailed information returned for single item query
type DataDetailResponse struct {
	ID          string                 `json:"id"`
	StationCode string                 `json:"station_code"`
	Parameter   string                 `json:"parameter"`
	Value       float64                `json:"value"`
	Unit        string                 `json:"unit"`
	Timestamp   time.Time              `json:"timestamp"`
	Status      string                 `json:"status"`
	RawPayload  map[string]interface{} `json:"raw_payload,omitempty"`
}
