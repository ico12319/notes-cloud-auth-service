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

type TodoClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewTodoClient(baseURL string) *TodoClient {
	return &TodoClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type TodoServiceError struct {
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func (e TodoServiceError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("todo service %d: %s", e.Status, e.Message)
	}

	return fmt.Sprintf("todo service returned status %d", e.Status)
}

func (c *TodoClient) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("todo service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var svcErr TodoServiceError
		svcErr.Status = resp.StatusCode

		if err := json.NewDecoder(resp.Body).Decode(&svcErr); err != nil {
			return fmt.Errorf("todo service returned status %d", resp.StatusCode)
		}

		if svcErr.Status == 0 {
			svcErr.Status = resp.StatusCode
		}

		return svcErr
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode todo service response: %w", err)
		}
	}

	return nil
}

func (c *TodoClient) jsonRequest(ctx context.Context, method string, url string, body any) (*http.Request, error) {
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

// --- Todo list endpoints ---

// CreateTodoList calls POST /api/v1/users/{userId}/todo-lists
func (c *TodoClient) CreateTodoList(
	ctx context.Context,
	userID string,
	body models.CreateTodoListRequest,
) (*models.TodoListView, error) {
	req, err := c.jsonRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/users/%s/todo-lists", c.baseURL, userID),
		body,
	)
	if err != nil {
		return nil, err
	}

	var view models.TodoListView
	return &view, c.do(req, &view)
}

// GetTodoListsWithTasks calls GET /api/v1/users/{userId}/todo-lists
func (c *TodoClient) GetTodoListsWithTasks(
	ctx context.Context,
	userID string,
) ([]models.TodoListWithTasksView, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%s/todo-lists", c.baseURL, userID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var views []models.TodoListWithTasksView
	return views, c.do(req, &views)
}

// GetTodoList calls GET /api/v1/users/{userId}/todo-lists/{listId}
func (c *TodoClient) GetTodoList(
	ctx context.Context,
	userID string,
	listID string,
) (*models.TodoListWithTasksView, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%s/todo-lists/%s", c.baseURL, userID, listID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var view models.TodoListWithTasksView
	return &view, c.do(req, &view)
}

// UpdateTodoList calls PUT /api/v1/users/{userId}/todo-lists/{listId}
func (c *TodoClient) UpdateTodoList(
	ctx context.Context,
	userID string,
	listID string,
	body models.UpdateTodoListRequest,
) (*models.TodoListView, error) {
	req, err := c.jsonRequest(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/users/%s/todo-lists/%s", c.baseURL, userID, listID),
		body,
	)
	if err != nil {
		return nil, err
	}

	var view models.TodoListView
	return &view, c.do(req, &view)
}

// DeleteTodoList calls DELETE /api/v1/users/{userId}/todo-lists/{listId}
func (c *TodoClient) DeleteTodoList(
	ctx context.Context,
	userID string,
	listID string,
) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/users/%s/todo-lists/%s", c.baseURL, userID, listID),
		nil,
	)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

// --- Todo task endpoints ---

// CreateTodoTask calls POST /api/v1/users/{userId}/todo-tasks
func (c *TodoClient) CreateTodoTask(
	ctx context.Context,
	userID string,
	body models.CreateTodoTaskRequest,
) (*models.TodoTaskView, error) {
	req, err := c.jsonRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/users/%s/todo-tasks", c.baseURL, userID),
		body,
	)
	if err != nil {
		return nil, err
	}

	var view models.TodoTaskView
	return &view, c.do(req, &view)
}

// GetStandaloneTasks calls GET /api/v1/users/{userId}/todo-tasks
func (c *TodoClient) GetStandaloneTasks(
	ctx context.Context,
	userID string,
) ([]models.TodoTaskView, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%s/todo-tasks", c.baseURL, userID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var views []models.TodoTaskView
	return views, c.do(req, &views)
}

// GetTodoTask calls GET /api/v1/users/{userId}/todo-tasks/{taskId}
func (c *TodoClient) GetTodoTask(
	ctx context.Context,
	userID string,
	taskID string,
) (*models.TodoTaskView, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/users/%s/todo-tasks/%s", c.baseURL, userID, taskID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	var view models.TodoTaskView
	return &view, c.do(req, &view)
}

// UpdateTodoTask calls PUT /api/v1/users/{userId}/todo-tasks/{taskId}
func (c *TodoClient) UpdateTodoTask(
	ctx context.Context,
	userID string,
	taskID string,
	body models.UpdateTodoTaskRequest,
) (*models.TodoTaskView, error) {
	req, err := c.jsonRequest(
		ctx,
		http.MethodPut,
		fmt.Sprintf("%s/api/v1/users/%s/todo-tasks/%s", c.baseURL, userID, taskID),
		body,
	)
	if err != nil {
		return nil, err
	}

	var view models.TodoTaskView
	return &view, c.do(req, &view)
}

// DeleteTodoTask calls DELETE /api/v1/users/{userId}/todo-tasks/{taskId}
func (c *TodoClient) DeleteTodoTask(
	ctx context.Context,
	userID string,
	taskID string,
) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		fmt.Sprintf("%s/api/v1/users/%s/todo-tasks/%s", c.baseURL, userID, taskID),
		nil,
	)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}