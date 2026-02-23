package invopop

import (
	"context"
	"fmt"
)

type AtPtService struct {
	client *Client
}

// NewAtPtService creates a new instance of the AtPtService.
func NewAtPtService(client *Client) *AtPtService {
	return &AtPtService{client: client}
}

// SetAtCredentialsRequest defines the body for the AT credentials upload API.
type SetAtCredentialsRequest struct {
	// Username is the username for the AT credentials.
	Username string `json:"username"`
	// Password is the password for the AT credentials.
	Password string `json:"password"`
}

// SetAtCredentials sets the at credentials for a specific suppliert in portugal.
//
// Endpoint: POST /apps/at-pt/v1/entry/{silo_entry_id}/credentials
func (s *AtPtService) SetAtCredentials(ctx context.Context, siloEntryID string, username, password string) error {
	if siloEntryID == "" {
		return fmt.Errorf("siloEntryID is required")
	}

	// 1. Prepare the request body
	reqBody := &SetAtCredentialsRequest{
		Username: username,
		Password: password,
	}

	// 2. Build the URL path as per documentation
	path := fmt.Sprintf("/apps/at-pt/v1/entry/%s/credentials", siloEntryID)

	// 3. Execute the request
	// Note: We use s.client.Post. The documentation specifies a 204 response (no body).
	// If your client's Post method expects a response object, you can pass nil.
	err := s.client.post(ctx, path, reqBody, nil)
	if err != nil {
		return fmt.Errorf("failed to upload peppol proof of ownership: %w", err)
	}

	return nil
}

type DocumentType string

// FT: Invoice (Fatura)
// FS: Simplified Invoice (Fatura Simplificada)
// FR: Invoice-Receipt (Fatura-Recibo)
// ND: Debit Note (Nota de Débito)
// NC: Credit Note (Nota de Crédito)
// RG: Receipt (Recibo)
// GR: Waybill (Guia de Remessa)
const (
	DocumentTypeInvoice           DocumentType = "FT"
	DocumentTypeSimplifiedInvoice DocumentType = "FS"
	DocumentTypeInvoiceReceipt    DocumentType = "FR"
	DocumentTypeDebitNote         DocumentType = "ND"
	DocumentTypeCreditNote        DocumentType = "NC"
	DocumentTypeReceipt           DocumentType = "RG"
	DocumentTypeWaybill           DocumentType = "GR"
)

// RegisterSeriesRequest defines the body for the series registration API.
type RegisterSeriesRequest struct {
	Base         string       `json:"base"`
	DocumentType DocumentType `json:"document_type"`
	Type         string       `json:"type"` // N (Normal), F (Formacao, Training), R (Recuperacao, Recovery)
}

type SeriesResponse struct {
	SeriesID       string       `json:"id"`
	Code           string       `json:"code"`
	Base           string       `json:"base"`
	DocumentType   DocumentType `json:"document_type"`
	Type           string       `json:"type"`
	ValidationCode string       `json:"validation_code"`
}

// RegisterSeries registers a new series for a specific supplier in Portugal.
//
// Endpoint: POST /apps/at-pt/v1/entry/{silo_entry_id}/series
func (s *AtPtService) RegisterSeries(ctx context.Context, siloEntryID string, base string, documentType DocumentType, seriesType string) (*SeriesResponse, error) {
	if siloEntryID == "" {
		return nil, fmt.Errorf("siloEntryID is required")
	}

	// 1. Prepare the request body by encoding the PDF bytes to Base64
	reqBody := &RegisterSeriesRequest{
		Base:         base,
		DocumentType: documentType,
		Type:         seriesType,
	}

	// 2. Build the URL path as per documentation
	path := fmt.Sprintf("/apps/at-pt/v1/entry/%s/series", siloEntryID)

	// 3. Execute the request
	// Note: We use s.client.Post.
	resp := new(SeriesResponse)
	err := s.client.post(ctx, path, reqBody, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to register series: %w", err)
	}

	return resp, nil
}

// GetSeries retrieves a series for a specific supplier in Portugal.
//
// Endpoint: GET /apps/at-pt/v1/entry/{silo_entry_id}/series/{series_id}
func (s *AtPtService) GetSeries(ctx context.Context, siloEntryID string, seriesID string) (*SeriesResponse, error) {
	if siloEntryID == "" {
		return nil, fmt.Errorf("siloEntryID is required")
	}
	if seriesID == "" {
		return nil, fmt.Errorf("seriesID is required")
	}

	// 2. Build the URL path as per documentation
	path := fmt.Sprintf("/apps/at-pt/v1/entry/%s/series/%s", siloEntryID, seriesID)

	// 3. Execute the request
	// Note: We use s.client.Get.
	resp := new(SeriesResponse)
	err := s.client.get(ctx, path, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get series: %w", err)
	}

	return resp, nil
}

type ListSeriesResponse struct {
	List []SeriesResponse `json:"list"`
}

// ListSeries lists all series for a specific supplier in Portugal.
//
// Endpoint: GET /apps/at-pt/v1/entry/{silo_entry_id}/series
func (s *AtPtService) ListSeries(ctx context.Context, siloEntryID string) (*ListSeriesResponse, error) {
	if siloEntryID == "" {
		return nil, fmt.Errorf("siloEntryID is required")
	}

	// 2. Build the URL path as per documentation
	path := fmt.Sprintf("/apps/at-pt/v1/entry/%s/series", siloEntryID)

	// 3. Execute the request
	resp := new(ListSeriesResponse)
	err := s.client.get(ctx, path, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list series: %w", err)
	}

	return resp, nil
}
