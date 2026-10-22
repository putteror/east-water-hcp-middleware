package thirdparty

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"eastwater-backend/internal/config"
	"eastwater-backend/internal/core/common"
	"eastwater-backend/internal/core/port"
	"eastwater-backend/internal/core/schema"
)

// Ensure Adapter implements ThirdPartyPort interface
var _ port.ThirdPartyPort = (*Adapter)(nil)

// Adapter implements port.ThirdPartyPort using standard HTTP client
type Adapter struct {
	cfg        config.ThirdPartyConfig
	httpClient *http.Client
}

// NewAdapter initializes a new 3rd-party HTTP client adapter
func NewAdapter(cfg config.ThirdPartyConfig) *Adapter {
	// Production-ready HTTP transport with connection pooling
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Adapter{
		cfg: cfg,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
}

// TransmitData sends payload to 3rd party ingestion API
func (a *Adapter) TransmitData(ctx context.Context, payload *schema.ThirdPartyIngestRequest) (*schema.ThirdPartyIngestResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/telemetry/ingest", strings.TrimRight(a.cfg.BaseURL, "/"))

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal 3rd party request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create 3rd party request: %w", err)
	}

	a.setHeaders(req)

	if a.cfg.Debug {
		log.Printf("[3rdParty] POST %s -> Payload: %s", endpoint, string(reqBody))
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, common.ErrThirdPartyTimeout
		}
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyService, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read 3rd party response: %w", err)
	}

	if a.cfg.Debug {
		log.Printf("[3rdParty] Status %d <- Response: %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a.handleErrorResponse(resp.StatusCode, respBody)
	}

	var ingestResp schema.ThirdPartyIngestResponse
	if err := json.Unmarshal(respBody, &ingestResp); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyBadResp, err)
	}

	return &ingestResp, nil
}

// FetchDataList queries list of records from 3rd party API
func (a *Adapter) FetchDataList(ctx context.Context, query schema.DataSearchQuery) (*schema.ThirdPartyQueryResponse, error) {
	baseURL := fmt.Sprintf("%s/api/v1/telemetry/records", strings.TrimRight(a.cfg.BaseURL, "/"))
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}

	q := u.Query()
	q.Set("page", strconv.Itoa(query.Page))
	q.Set("limit", strconv.Itoa(query.Limit))
	if query.StationCode != "" {
		q.Set("station_code", query.StationCode)
	}
	if query.Parameter != "" {
		q.Set("parameter", query.Parameter)
	}
	if query.StartDate != "" {
		q.Set("start_date", query.StartDate)
	}
	if query.EndDate != "" {
		q.Set("end_date", query.EndDate)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build 3rd party query request: %w", err)
	}

	a.setHeaders(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, common.ErrThirdPartyTimeout
		}
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyService, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read 3rd party response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a.handleErrorResponse(resp.StatusCode, respBody)
	}

	var queryResp schema.ThirdPartyQueryResponse
	if err := json.Unmarshal(respBody, &queryResp); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyBadResp, err)
	}

	return &queryResp, nil
}

// FetchDataByID queries single record detail from 3rd party API
func (a *Adapter) FetchDataByID(ctx context.Context, id string) (*schema.ThirdPartyItemData, error) {
	endpoint := fmt.Sprintf("%s/api/v1/telemetry/records/%s", strings.TrimRight(a.cfg.BaseURL, "/"), url.PathEscape(id))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build 3rd party request: %w", err)
	}

	a.setHeaders(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, common.ErrThirdPartyTimeout
		}
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyService, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read 3rd party response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, common.ErrNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, a.handleErrorResponse(resp.StatusCode, respBody)
	}

	var item schema.ThirdPartyItemData
	if err := json.Unmarshal(respBody, &item); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrThirdPartyBadResp, err)
	}

	return &item, nil
}

// CheckHealth verifies connectivity with 3rd party API
func (a *Adapter) CheckHealth(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/health", strings.TrimRight(a.cfg.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	a.setHeaders(req)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("third party health check returned status %d", resp.StatusCode)
	}
	return nil
}

func (a *Adapter) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if a.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.cfg.APIKey)
		req.Header.Set("x-api-key", a.cfg.APIKey)
	}
}

func (a *Adapter) handleErrorResponse(statusCode int, body []byte) error {
	var errObj struct {
		Message   string `json:"message"`
		ErrorCode string `json:"error_code"`
	}
	_ = json.Unmarshal(body, &errObj)

	msg := errObj.Message
	if msg == "" {
		msg = string(body)
	}

	return &common.ThirdPartyError{
		StatusCode: statusCode,
		ErrorCode:  errObj.ErrorCode,
		Message:    msg,
		RawBody:    string(body),
	}
}
