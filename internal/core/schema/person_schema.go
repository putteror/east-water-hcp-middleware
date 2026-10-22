package schema

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"
)

// FlexStringSlice supports unmarshaling from []string, []int, raw string, or single number
type FlexStringSlice []string

func (f *FlexStringSlice) UnmarshalJSON(data []byte) error {
	var rawList []interface{}
	if err := json.Unmarshal(data, &rawList); err == nil {
		for _, item := range rawList {
			switch v := item.(type) {
			case string:
				if trimmed := strings.TrimSpace(v); trimmed != "" {
					*f = append(*f, trimmed)
				}
			case float64:
				*f = append(*f, fmt.Sprintf("%.0f", v))
			}
		}
		return nil
	}

	var singleStr string
	if err := json.Unmarshal(data, &singleStr); err == nil {
		for _, p := range strings.Split(singleStr, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				*f = append(*f, trimmed)
			}
		}
		return nil
	}

	var singleNum float64
	if err := json.Unmarshal(data, &singleNum); err == nil {
		*f = append(*f, fmt.Sprintf("%.0f", singleNum))
		return nil
	}

	return nil
}

func (f FlexStringSlice) ToStringSlice() []string {
	if len(f) == 0 {
		return nil
	}
	return []string(f)
}

// CreatePersonRequest maps JSON payload (or multipart/form-data) received from VMS software
type CreatePersonRequest struct {
	PersonName          string                  `json:"personName" form:"personName" binding:"required" example:"Somchai Jaidee"`
	PersonCode          string                  `json:"personCode" form:"personCode" binding:"required" example:"P001"`
	LicensePlateTexts   []string              `json:"licensePlateTexts,omitempty" form:"licensePlateTexts" example:"1กก1234,2ขข5678"`
	LicensePlateTextRaw string                `json:"licensePlateText,omitempty" form:"licensePlateText" swaggerignore:"true"` // Fallback for single or comma-separated value
	// DeviceNames: Doors/devices to grant access (Specify either deviceNames OR accessLevelIds). Supports partial match (e.g. "Main Entrance")
	DeviceNames         []string              `json:"deviceNames,omitempty" form:"deviceNames" example:"Main Entrance,Building A"`
	DeviceNameRaw       string                `json:"deviceName,omitempty" form:"deviceName" swaggerignore:"true"`
	// AccessLevelIDs: HikCentral Access Level IDs (Specify either accessLevelIds OR deviceNames). Accepts array of strings or numbers (e.g. ["1", "2"] or [1, 2])
	AccessLevelIDs      FlexStringSlice       `json:"accessLevelIds,omitempty" form:"accessLevelIds" example:"1,2"`
	AccessLevelIDRaw    string                `json:"accessLevelId,omitempty" form:"accessLevelId" swaggerignore:"true"`
	StartDateTime       string                `json:"startDateTime,omitempty" form:"startDateTime" example:"2026-01-01T00:00:00Z"`
	EndDateTime         string                `json:"endDateTime,omitempty" form:"endDateTime" example:"2026-12-31T23:59:59Z"`
	FaceImage           string                `json:"faceImage,omitempty" form:"faceImageBase64" example:"/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."` // Base64 JPEG/PNG string
	FaceImageFile       *multipart.FileHeader `json:"-" form:"faceImage" swaggerignore:"true"`                                                                   // Multipart upload fallback
}

// GetFaceImageFile retrieves the primary uploaded face image file
func (r *CreatePersonRequest) GetFaceImageFile() *multipart.FileHeader {
	return r.FaceImageFile
}

// GetAccessLevelIDs extracts access level IDs whether passed as array or comma-separated string
func (r *CreatePersonRequest) GetAccessLevelIDs() []string {
	if len(r.AccessLevelIDs) > 0 {
		return r.AccessLevelIDs.ToStringSlice()
	}
	if r.AccessLevelIDRaw != "" {
		parts := strings.Split(r.AccessLevelIDRaw, ",")
		var res []string
		for _, p := range parts {
			if val := strings.TrimSpace(p); val != "" {
				res = append(res, val)
			}
		}
		return res
	}
	return nil
}

// GetDeviceNames extracts device names whether passed as array or comma-separated string
func (r *CreatePersonRequest) GetDeviceNames() []string {
	var results []string
	for _, d := range r.DeviceNames {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			parts := strings.Split(trimmed, ",")
			for _, p := range parts {
				if val := strings.TrimSpace(p); val != "" {
					results = append(results, val)
				}
			}
		}
	}
	if len(results) == 0 && r.DeviceNameRaw != "" {
		parts := strings.Split(r.DeviceNameRaw, ",")
		for _, p := range parts {
			if val := strings.TrimSpace(p); val != "" {
				results = append(results, val)
			}
		}
	}
	return results
}

// GetLicensePlateList extracts license plates whether sent as array or comma-separated string
func (r *CreatePersonRequest) GetLicensePlateList() []string {
	var results []string
	for _, lp := range r.LicensePlateTexts {
		trimmed := strings.TrimSpace(lp)
		if trimmed != "" {
			// In case it's comma-separated inside one element
			parts := strings.Split(trimmed, ",")
			for _, p := range parts {
				if val := strings.TrimSpace(p); val != "" {
					results = append(results, val)
				}
			}
		}
	}
	if len(results) == 0 && r.LicensePlateTextRaw != "" {
		parts := strings.Split(r.LicensePlateTextRaw, ",")
		for _, p := range parts {
			if val := strings.TrimSpace(p); val != "" {
				results = append(results, val)
			}
		}
	}
	return results
}

// DeletePeopleRequest represents the payload for deleting people by personCode list
type DeletePeopleRequest struct {
	PersonCodes []string `json:"personCodes" binding:"required,min=1" example:"P001,P002"`
}

// PersonSearchQuery maps query parameters for listing people
type PersonSearchQuery struct {
	PaginationQuery
	PersonName string `form:"personName"`
	PersonCode string `form:"personCode"`
	Code       string `form:"code"` // Alias for personCode
}

// SetDefaults applies pagination defaults and aliases code to personCode
func (q *PersonSearchQuery) SetDefaults() {
	q.PaginationQuery.SetDefaults()
	if q.PersonCode == "" && q.Code != "" {
		q.PersonCode = q.Code
	}
}

// PersonBasicInfo represents lightweight person record returned in list response
type PersonBasicInfo struct {
	PersonCode        string   `json:"personCode" example:"P001"`
	PersonName        string   `json:"personName" example:"Somchai Jaidee"`
	LicensePlateTexts []string `json:"licensePlateTexts,omitempty" example:"1กก1234"`
	DeviceNames       []string `json:"deviceNames,omitempty" example:"Main Entrance Door 1"`
	AccessLevelIDs    []string `json:"accessLevelIds,omitempty" example:"1,2"`
	StartDateTime     string   `json:"startDateTime,omitempty" example:"2026-01-01T00:00:00Z"`
	EndDateTime       string   `json:"endDateTime,omitempty" example:"2026-12-31T23:59:59Z"`
	HasFaceImage      bool     `json:"hasFaceImage" example:"true"`
	Status            string   `json:"status" example:"ACTIVE"`
}

// PersonResponse represents full response when creating, updating, or retrieving a person
type PersonResponse struct {
	Action            string    `json:"action,omitempty" example:"created"` // "created" or "updated"
	PersonCode        string    `json:"personCode" example:"P001"`
	PersonName        string    `json:"personName" example:"Somchai Jaidee"`
	LicensePlateTexts []string  `json:"licensePlateTexts,omitempty" example:"1กก1234"`
	DeviceNames       []string  `json:"deviceNames,omitempty" example:"Main Entrance Door 1,Building A Turnstile"`
	AccessLevelIDs    []string  `json:"accessLevelIds,omitempty" example:"1,2"`
	StartDateTime     string    `json:"startDateTime,omitempty" example:"2026-01-01T00:00:00Z"`
	EndDateTime       string    `json:"endDateTime,omitempty" example:"2026-12-31T23:59:59Z"`
	HasFaceImage      bool      `json:"hasFaceImage" example:"true"`
	FaceImage         string    `json:"faceImage,omitempty" example:"/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."` // Single Base64 JPEG string
	Status            string    `json:"status" example:"ACTIVE"`
	ProcessedAt       time.Time `json:"processedAt"`
}

// UpdatePersonRequest maps JSON payload (or multipart/form-data) for updating an existing person
type UpdatePersonRequest struct {
	PersonName          string                `json:"personName,omitempty" form:"personName" example:"Somchai Updated"`
	LicensePlateTexts   []string              `json:"licensePlateTexts,omitempty" form:"licensePlateTexts" example:"9กก9999"`
	LicensePlateTextRaw string                `json:"licensePlateText,omitempty" form:"licensePlateText" swaggerignore:"true"`
	// DeviceNames: Doors/devices to grant access (Specify either deviceNames OR accessLevelIds). Supports partial match (e.g. "Main Entrance")
	DeviceNames         []string              `json:"deviceNames,omitempty" form:"deviceNames" example:"Main Entrance,Building A"`
	DeviceNameRaw       string                `json:"deviceName,omitempty" form:"deviceName" swaggerignore:"true"`
	// AccessLevelIDs: HikCentral Access Level IDs (Specify either accessLevelIds OR deviceNames). Accepts array of strings or numbers (e.g. ["1", "2"] or [1, 2])
	AccessLevelIDs      FlexStringSlice       `json:"accessLevelIds,omitempty" form:"accessLevelIds" example:"1,2"`
	AccessLevelIDRaw    string                `json:"accessLevelId,omitempty" form:"accessLevelId" swaggerignore:"true"`
	StartDateTime       string                `json:"startDateTime,omitempty" form:"startDateTime" example:"2026-01-01T00:00:00Z"`
	EndDateTime         string                `json:"endDateTime,omitempty" form:"endDateTime" example:"2026-12-31T23:59:59Z"`
	FaceImage           string                `json:"faceImage,omitempty" form:"faceImageBase64" example:"/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP...mockBase64..."` // Updated Base64 JPEG/PNG string
	FaceImageFile       *multipart.FileHeader `json:"-" form:"faceImage" swaggerignore:"true"`                                                           // Multipart upload fallback
}

// GetFaceImageFile retrieves the primary uploaded face image file
func (r *UpdatePersonRequest) GetFaceImageFile() *multipart.FileHeader {
	return r.FaceImageFile
}

// GetAccessLevelIDs extracts access level IDs whether passed as array or comma-separated string
func (r *UpdatePersonRequest) GetAccessLevelIDs() []string {
	if len(r.AccessLevelIDs) > 0 {
		return r.AccessLevelIDs.ToStringSlice()
	}
	if r.AccessLevelIDRaw != "" {
		parts := strings.Split(r.AccessLevelIDRaw, ",")
		var res []string
		for _, p := range parts {
			if val := strings.TrimSpace(p); val != "" {
				res = append(res, val)
			}
		}
		return res
	}
	return nil
}

// GetDeviceNames extracts device names whether passed as array or comma-separated string
func (r *UpdatePersonRequest) GetDeviceNames() []string {
	var results []string
	for _, d := range r.DeviceNames {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			parts := strings.Split(trimmed, ",")
			for _, p := range parts {
				if val := strings.TrimSpace(p); val != "" {
					results = append(results, val)
				}
			}
		}
	}
	if len(results) == 0 && r.DeviceNameRaw != "" {
		parts := strings.Split(r.DeviceNameRaw, ",")
		for _, p := range parts {
			if val := strings.TrimSpace(p); val != "" {
				results = append(results, val)
			}
		}
	}
	return results
}

func (r *UpdatePersonRequest) GetLicensePlateList() []string {
	var results []string
	for _, lp := range r.LicensePlateTexts {
		trimmed := strings.TrimSpace(lp)
		if trimmed != "" {
			parts := strings.Split(trimmed, ",")
			for _, p := range parts {
				if val := strings.TrimSpace(p); val != "" {
					results = append(results, val)
				}
			}
		}
	}
	if len(results) == 0 && r.LicensePlateTextRaw != "" {
		parts := strings.Split(r.LicensePlateTextRaw, ",")
		for _, p := range parts {
			if val := strings.TrimSpace(p); val != "" {
				results = append(results, val)
			}
		}
	}
	return results
}

// PersonDeleteResponse represents result of batch delete operation
type PersonDeleteResponse struct {
	TotalRequested int      `json:"totalRequested"`
	DeletedCount   int      `json:"deletedCount"`
	DeletedCodes   []string `json:"deletedCodes"`
	FailedCodes    []string `json:"failedCodes,omitempty"`
}

// DefaultKnownDevices provides registered doors, gates, and access points in the facility
var DefaultKnownDevices = []string{
	"Main Entrance Door 1",
	"Main Entrance Turnstile",
	"Building A Entrance",
	"Building A Turnstile",
	"Building B Entrance",
	"East Water Gate Barrier In",
	"East Water Gate Barrier Out",
	"Pump Station Door 1",
	"Warehouse Security Door",
}

// MatchDeviceNames resolves input device names against known devices using case-insensitive partial/substring matching.
// - If an input matches one or more known devices (e.g. "Main" matches "Main Entrance Door 1"), all matching full names are returned.
// - If an input does not match any known device, the trimmed input is preserved as-is.
// - Output is deduplicated and order-preserved.
func MatchDeviceNames(inputNames []string, knownDevices []string) []string {
	if len(inputNames) == 0 {
		return nil
	}
	if len(knownDevices) == 0 {
		knownDevices = DefaultKnownDevices
	}

	seen := make(map[string]bool)
	var resolved []string

	for _, input := range inputNames {
		trimmed := strings.TrimSpace(input)
		if trimmed == "" {
			continue
		}
		lowerInput := strings.ToLower(trimmed)

		matchedCount := 0
		for _, known := range knownDevices {
			if strings.Contains(strings.ToLower(known), lowerInput) {
				if !seen[known] {
					seen[known] = true
					resolved = append(resolved, known)
				}
				matchedCount++
			}
		}

		if matchedCount == 0 {
			if !seen[trimmed] {
				seen[trimmed] = true
				resolved = append(resolved, trimmed)
			}
		}
	}

	return resolved
}
