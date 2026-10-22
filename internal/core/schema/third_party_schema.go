package schema

// ThirdPartyIngestRequest is the payload format expected by the 3rd party external API
type ThirdPartyIngestRequest struct {
	DeviceID   string                 `json:"device_id"`
	MetricName string                 `json:"metric_name"`
	MetricVal  float64                `json:"metric_val"`
	MetricUnit string                 `json:"metric_unit"`
	Timestamp  string                 `json:"timestamp"`
	Tags       map[string]interface{} `json:"tags,omitempty"`
}

// ThirdPartyIngestResponse is the response returned by the 3rd party API
type ThirdPartyIngestResponse struct {
	AckCode   string `json:"ack_code"`
	TicketID  string `json:"ticket_id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// ThirdPartyQueryResponse is the list/query response returned by the 3rd party API
type ThirdPartyQueryResponse struct {
	TotalCount int                  `json:"total_count"`
	Items      []ThirdPartyItemData `json:"items"`
}

// ThirdPartyItemData represents a record in the 3rd party response
type ThirdPartyItemData struct {
	RecordID   string                 `json:"record_id"`
	DeviceID   string                 `json:"device_id"`
	MetricName string                 `json:"metric_name"`
	Value      float64                `json:"value"`
	Unit       string                 `json:"unit"`
	OccurredAt string                 `json:"occurred_at"`
	State      string                 `json:"state"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}
