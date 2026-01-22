//  See here for more info: https://docs.invopop.com/api-ref/apps/peppol/ownership-upload.
// This file implements `POST https://api.invopop.com/v1/apps/peppol/ownership/upload` API.
// https://api.invopop.com/apps/peppol/v1/entry/{silo_entry_id}/ownership

package invopop

import (
	"context"
	"encoding/base64"
	"fmt"
)

// PeppolOwnershipRequest defines the body for the ownership upload API.
// The API requires the binary data to be Base64 encoded.
type PeppolOwnershipRequest struct {
	// Data is the Base64 encoded binary data of the ownership document (PDF).
	Data string `json:"data"`
}

// UploadPeppolOwnership uploads a proof of ownership PDF for a specific silo entry.
// This is required for Peppol participant registration compliance.
//
// Endpoint: POST /apps/peppol/v1/entry/{silo_entry_id}/ownership
func (c *Client) UploadPeppolOwnership(ctx context.Context, siloEntryID string, pdfData []byte) error {
	if siloEntryID == "" {
		return fmt.Errorf("siloEntryID is required")
	}

	// 1. Prepare the request body by encoding the PDF bytes to Base64
	reqBody := &PeppolOwnershipRequest{
		Data: base64.StdEncoding.EncodeToString(pdfData),
	}

	// 2. Build the URL path as per documentation
	path := fmt.Sprintf("/apps/peppol/v1/entry/%s/ownership", siloEntryID)

	// 3. Execute the request
	// Note: We use c.Post. The documentation specifies a 204 response (no body).
	// If your client's Post method expects a response object, you can pass nil.
	err := c.post(ctx, path, reqBody, nil)
	if err != nil {
		return fmt.Errorf("failed to upload peppol proof of ownership: %w", err)
	}

	return nil
}
