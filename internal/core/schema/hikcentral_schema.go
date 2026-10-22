package schema

// HikApiResponse is the standard envelope returned by HikCentral Pro Artemis OpenAPI
type HikApiResponse[T any] struct {
	Code string `json:"code"` // "0" indicates success in HikCentral
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

// HikCreatePersonPayload is our internal payload passed from Service to Port
type HikCreatePersonPayload struct {
	AppointRecordID string           `json:"appointRecordId,omitempty"`
	PersonCode      string           `json:"personCode"`
	PersonName      string           `json:"personName"`
	BeginTime       string           `json:"beginTime,omitempty"`
	EndTime         string           `json:"endTime,omitempty"`
	Vehicles        []HikVehicleData `json:"vehicles,omitempty"`
	FaceBase64      string           `json:"faceBase64,omitempty"`  // Single Base64 JPEG string
	FaceBase64s     []string         `json:"faceBase64s,omitempty"` // Fallback/compat
	DeviceNames     []string         `json:"deviceNames,omitempty"` // Resolved accessible doors / devices
	AccessLevelIDs  []string         `json:"accessLevelIds,omitempty"` // HikCentral Access Level IDs
}

// GetFaceBase64 retrieves the single face Base64 string
func (p *HikCreatePersonPayload) GetFaceBase64() string {
	if p.FaceBase64 != "" {
		return p.FaceBase64
	}
	if len(p.FaceBase64s) > 0 {
		return p.FaceBase64s[0]
	}
	return ""
}

// HikVehicleData represents vehicle plate info
type HikVehicleData struct {
	PlateNo string `json:"plateNo"`
}

// HikPersonItem represents visitor/person details in our port contract
type HikPersonItem struct {
	AppointRecordID string           `json:"appointRecordId,omitempty"`
	PersonID        string           `json:"personId,omitempty"`
	PersonCode      string           `json:"personCode"`
	PersonName      string           `json:"personName"`
	BeginTime       string           `json:"beginTime,omitempty"`
	EndTime         string           `json:"endTime,omitempty"`
	Vehicles        []HikVehicleData `json:"vehicles,omitempty"`
	DeviceNames     []string         `json:"deviceNames,omitempty"`
	AccessLevelIDs  []string         `json:"accessLevelIds,omitempty"`
	HasFaceImage    bool             `json:"hasFaceImage"`
	FaceBase64      string           `json:"faceBase64,omitempty"`  // Single Base64 JPEG string
	FacesCount      int              `json:"facesCount"`
	FaceBase64s     []string         `json:"faceBase64s,omitempty"`
	Status          string           `json:"status"`
}

// GetFaceBase64 retrieves the single face Base64 string
func (i *HikPersonItem) GetFaceBase64() string {
	if i.FaceBase64 != "" {
		return i.FaceBase64
	}
	if len(i.FaceBase64s) > 0 {
		return i.FaceBase64s[0]
	}
	return ""
}

// --- Official HikCentral Pro Visitor V2 / Artemis API Schemas ---

// HikVisitorAppointmentV2Request maps to POST /artemis/api/visitor/v2/appointment
type HikVisitorAppointmentV2Request struct {
	AppointRecordID  string                    `json:"appointRecordId,omitempty"` // For update
	ReceptionistID   string                    `json:"receptionistId,omitempty"`
	AppointStartTime string                    `json:"appointStartTime"` // ISO 8601 e.g. "2021-05-09T15:00:00+08:00"
	AppointEndTime   string                    `json:"appointEndTime"`
	VisitReasonType  int                       `json:"visitReasonType"` // 0-business, 1-training, 2-visit, 3-meeting, 4-others
	AccessInfo       *HikAppointmentAccessInfo `json:"accessInfo,omitempty"`
	VisitorInfoList  []HikVisitorInfoWrapper   `json:"visitorInfoList"`
}

type HikAppointmentAccessInfo struct {
	AccessLevelList []HikAppointmentAccessLevelWrapper `json:"accessLevelList,omitempty"`
}

type HikAppointmentAccessLevelWrapper struct {
	AccessLevel HikAppointmentAccessLevel `json:"accessLevel"`
}

type HikAppointmentAccessLevel struct {
	ID       int                                `json:"id"`
	BaseInfo *HikAppointmentAccessLevelBaseInfo `json:"baseInfo,omitempty"`
}

type HikAppointmentAccessLevelBaseInfo struct {
	Name string `json:"name,omitempty"`
}

type HikVisitorInfoWrapper struct {
	VisitorInfo HikVisitorInfo `json:"VisitorInfo"`
}

type HikVisitorInfo struct {
	VisitorID         string            `json:"visitorId,omitempty"`
	VisitorGivenName  string            `json:"visitorGivenName"`
	VisitorFamilyName string            `json:"visitorFamilyName"`
	Gender            int               `json:"gender"` // 0-unknown, 1-male, 2-female
	PlateNo           string            `json:"plateNo,omitempty"`
	CertificateType   int               `json:"certificateType,omitempty"` // 111-ID card
	CertificateNo     string            `json:"certificateNo,omitempty"`   // We map personCode here
	Remark            string            `json:"remark,omitempty"`
	Faces             []HikFaceDataItem `json:"faces,omitempty"`
}

type HikFaceDataItem struct {
	FaceData string `json:"faceData"` // Base64 encoded JPEG
}

// HikAppointmentV2Result is the data payload returned by /visitor/v2/appointment
type HikAppointmentV2Result struct {
	AppointRecordID string `json:"appointRecordId"`
	VisitorID       string `json:"visitorId"`
	AppointCode     string `json:"appointCode"`
	QRCodeImage     string `json:"qrCodeImage,omitempty"`
}

// HikAppointmentListRequest maps to POST /artemis/api/visitor/v1/appointment/appointmentlist
type HikAppointmentListRequest struct {
	PageNo           int    `json:"pageNo"`
	PageSize         int    `json:"pageSize"`
	AppointStartTime string `json:"appointStartTime"`
	AppointEndTime   string `json:"appointEndTime"`
	VisitorName      string `json:"visitorName,omitempty"`
	IdentiCode       string `json:"identiCode,omitempty"` // Matches certificateNo / personCode
	AppointState     int    `json:"appointState"`         // -1 (all)
}

// HikAppointmentListResult is returned by /appointmentlist
type HikAppointmentListResult struct {
	Total    int                      `json:"total"`
	PageNo   int                      `json:"pageNo"`
	PageSize int                      `json:"pageSize"`
	List     []HikAppointmentListItem `json:"list"`
}

type HikAppointmentListItem struct {
	AppointID        string                     `json:"appointID"`
	AppointCode      string                     `json:"appointCode"`
	AppointStartTime string                     `json:"appointStartTime"`
	AppointEndTime   string                     `json:"appointEndTime"`
	AppointStatus    string                     `json:"appointStatus"`
	VisitorInfo      HikAppointmentItemVisitor  `json:"visitorInfo"`
}

type HikAppointmentItemVisitor struct {
	VisitorID         string `json:"visitorId"`
	VisitorGivenName  string `json:"visitorGivenName"`
	VisitorFamilyName string `json:"visitorFamilyName"`
	PlateNo           string `json:"plateNo,omitempty"`
	CertificateNo     string `json:"certificateNo,omitempty"`
	HaveVisitorHeadPic int   `json:"haveVisitorHeadPic"`
}
