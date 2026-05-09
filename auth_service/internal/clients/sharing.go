package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

type SharingClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewSharingClient(baseURL string) *SharingClient {
	return &SharingClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type SharingServiceError struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"error"`
	Message      string `json:"message"`
}

func (e SharingServiceError) Error() string {
	return fmt.Sprintf("sharing service %d %s: %s", e.Status, e.ErrorMessage, e.Message)
}

func (c *SharingClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sharing service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var svcErr SharingServiceError
		svcErr.Status = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&svcErr)
		return svcErr
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

func (c *SharingClient) jsonPost(ctx context.Context, url string, body any) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// --- Share link endpoints ---

// CreateShareLink calls POST /api/v1/users/{userId}/notes/{noteId}/share-links
//
// The sharing-service does not require a request body for this endpoint.
// It creates a VIEW-only share link that expires automatically.
func (c *SharingClient) CreateShareLink(
	ctx context.Context,
	userID string,
	noteID string,
) (*models.ShareLinkView, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/users/%s/notes/%s/share-links", c.baseURL, userID, noteID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var view models.ShareLinkView
	return &view, c.do(req, &view)
}

// OpenShareLink calls GET /api/v1/share-links/{token}
//
// This validates the token in sharing-service, then sharing-service calls notes-service
// and returns the shared note.
func (c *SharingClient) OpenShareLink(
	ctx context.Context,
	token string,
) (*models.SharedNoteView, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/share-links/%s", c.baseURL, token),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var view models.SharedNoteView
	return &view, c.do(req, &view)
}