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

type notesClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewNotesClient(baseURL string) *notesClient {
	return &notesClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type notesServiceError struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
}

func (e *notesServiceError) Err() error {
	return fmt.Errorf("notes service %d: %s", e.Status, e.Title)
}

func (c *notesClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notes service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var svcErr notesServiceError
		svcErr.Status = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&svcErr)
		return svcErr.Err()
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *notesClient) jsonRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// GetAll calls GET /api/users/{userId}/notes
func (c *notesClient) GetAll(ctx context.Context, userID string) ([]models.NoteView, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/users/%s/notes", c.baseURL, userID),
		nil)
	if err != nil {
		return nil, err
	}
	var notes []models.NoteView
	return notes, c.do(req, &notes)
}

// GetByID calls GET /api/users/{userId}/notes/{noteId}
func (c *notesClient) GetByID(ctx context.Context, userID, noteID string) (*models.NoteView, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/users/%s/notes/%s", c.baseURL, userID, noteID),
		nil)
	if err != nil {
		return nil, err
	}
	var note models.NoteView
	return &note, c.do(req, &note)
}

// Create calls POST /api/users/{userId}/notes
func (c *notesClient) Create(ctx context.Context, userID string, body models.NoteRequest) (*models.NoteView, error) {
	req, err := c.jsonRequest(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/users/%s/notes", c.baseURL, userID),
		body)
	if err != nil {
		return nil, err
	}
	var note models.NoteView
	return &note, c.do(req, &note)
}

// Update calls PUT /api/users/{userId}/notes/{noteId}
func (c *notesClient) Update(ctx context.Context, userID, noteID string, body models.NoteRequest) (*models.NoteView, error) {
	req, err := c.jsonRequest(ctx, http.MethodPut,
		fmt.Sprintf("%s/api/users/%s/notes/%s", c.baseURL, userID, noteID),
		body)
	if err != nil {
		return nil, err
	}
	var note models.NoteView
	return &note, c.do(req, &note)
}

// Delete calls DELETE /api/users/{userId}/notes/{noteId}
func (c *notesClient) Delete(ctx context.Context, userID, noteID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/users/%s/notes/%s", c.baseURL, userID, noteID),
		nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}