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

type ReminderNotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewReminderNotificationClient(baseURL string) *ReminderNotificationClient {
	return &ReminderNotificationClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type ReminderServiceError struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"error"`
	Message      string `json:"message"`
}

func (e ReminderServiceError) Error() string {
	return fmt.Sprintf("reminder service %d %s: %s", e.Status, e.ErrorMessage, e.Message)
}

func (c *ReminderNotificationClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reminder service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var svcErr ReminderServiceError
		svcErr.Status = resp.StatusCode
		_ = json.NewDecoder(resp.Body).Decode(&svcErr)
		return svcErr
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *ReminderNotificationClient) jsonPost(ctx context.Context, url string, body any) (*http.Request, error) {
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

// --- Reminder endpoints ---

// CreateReminder calls POST /api/users/{userId}/reminders
func (c *ReminderNotificationClient) CreateReminder(ctx context.Context, userID string, body models.ReminderRequest) (*models.ReminderView, error) {
	req, err := c.jsonPost(ctx, fmt.Sprintf("%s/api/users/%s/reminders", c.baseURL, userID), body)
	if err != nil {
		return nil, err
	}
	var view models.ReminderView
	return &view, c.do(req, &view)
}

// UpdateReminder calls PUT /api/users/{userId}/reminders
func (c *ReminderNotificationClient) UpdateReminder(ctx context.Context, userID string, body models.ReminderRequest) (*models.ReminderView, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/api/users/%s/reminders", c.baseURL, userID),
		bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var view models.ReminderView
	return &view, c.do(req, &view)
}

// GetReminders calls GET /api/users/{userId}/reminders
// filter: "" = all, "PENDING" or "COMPLETED" to filter by status
func (c *ReminderNotificationClient) GetReminders(ctx context.Context, userID string, filter string) ([]models.ReminderView, error) {
	url := fmt.Sprintf("%s/api/users/%s/reminders", c.baseURL, userID)
	if filter != "" {
		url += "?status=" + filter
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var views []models.ReminderView
	return views, c.do(req, &views)
}

// GetReminderByID calls GET /api/users/{userId}/reminders/{id}
func (c *ReminderNotificationClient) GetReminderByID(ctx context.Context, userID, id string) (*models.ReminderView, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/users/%s/reminders/%s", c.baseURL, userID, id),
		nil)
	if err != nil {
		return nil, err
	}
	var view models.ReminderView
	return &view, c.do(req, &view)
}

// DeleteReminder calls DELETE /api/users/{userId}/reminders/{id}
func (c *ReminderNotificationClient) DeleteReminder(ctx context.Context, userID, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/users/%s/reminders/%s", c.baseURL, userID, id),
		nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// --- Notification endpoints ---

// GetNotifications calls GET /api/users/{userId}/notifications
// read: nil = all, false = unread only, true = read only
func (c *ReminderNotificationClient) GetNotifications(ctx context.Context, userID string, read *bool) ([]models.NotificationView, error) {
	url := fmt.Sprintf("%s/api/users/%s/notifications", c.baseURL, userID)
	if read != nil {
		if *read {
			url += "?read=true"
		} else {
			url += "?read=false"
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var views []models.NotificationView
	return views, c.do(req, &views)
}

// CountUnreadNotifications calls GET /api/users/{userId}/notifications/unread-count
func (c *ReminderNotificationClient) CountUnreadNotifications(ctx context.Context, userID string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/users/%s/notifications/unread-count", c.baseURL, userID),
		nil)
	if err != nil {
		return 0, err
	}
	var count int64
	return count, c.do(req, &count)
}

// MarkNotificationAsRead calls POST /api/users/{userId}/notifications/{id}/read
func (c *ReminderNotificationClient) MarkNotificationAsRead(ctx context.Context, userID, id string) (*models.NotificationView, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/users/%s/notifications/%s/read", c.baseURL, userID, id),
		nil)
	if err != nil {
		return nil, err
	}
	var view models.NotificationView
	return &view, c.do(req, &view)
}

// MarkAllNotificationsAsRead calls POST /api/users/{userId}/notifications/read-all
func (c *ReminderNotificationClient) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/users/%s/notifications/read-all", c.baseURL, userID),
		nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// DeleteAllNotifications calls DELETE /api/users/{userId}/notifications
func (c *ReminderNotificationClient) DeleteAllNotifications(ctx context.Context, userID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/api/users/%s/notifications", c.baseURL, userID),
		nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}
