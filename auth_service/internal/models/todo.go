package models

import "time"

type TodoPriority string

const (
	TodoPriorityLow    TodoPriority = "LOW"
	TodoPriorityMedium TodoPriority = "MEDIUM"
	TodoPriorityHigh   TodoPriority = "HIGH"
)

type CreateTodoListRequest struct {
	Title string `json:"title"`
}

type UpdateTodoListRequest struct {
	Title string `json:"title"`
}

type TodoListView struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TodoListWithTasksView struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Tasks     []TodoTaskView `json:"tasks"`
}

type CreateTodoTaskRequest struct {
	ListID   *string      `json:"listId,omitempty"`
	Title    string       `json:"title"`
	Priority TodoPriority `json:"priority"`
	DueDate  *time.Time   `json:"dueDate,omitempty"`
}

type UpdateTodoTaskRequest struct {
	Title    *string       `json:"title,omitempty"`
	Priority *TodoPriority `json:"priority,omitempty"`
	DueDate  *time.Time    `json:"dueDate,omitempty"`
	Done     *bool         `json:"done,omitempty"`
}

type TodoTaskView struct {
	ID        string       `json:"id"`
	ListID    *string      `json:"listId"`
	UserID    string       `json:"userId"`
	Title     string       `json:"title"`
	Done      bool         `json:"done"`
	Priority  TodoPriority `json:"priority"`
	DueDate   *time.Time   `json:"dueDate"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}