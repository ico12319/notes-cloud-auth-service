package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/clients"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
	"log"
	"net/http"
	"time"
)

type ctxWithUserMetadata struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	userID     string
}

type handler struct {
	todoClient     *clients.TodoClient
	notesClient    *clients.NotesClient
	sharingClient  *clients.SharingClient
	reminderClient *clients.ReminderNotificationClient
}

func NewHandler(todoClient *clients.TodoClient, notesClient *clients.NotesClient, sharingClient *clients.SharingClient, reminderClient *clients.ReminderNotificationClient) *handler {
	return &handler{
		todoClient:     todoClient,
		notesClient:    notesClient,
		sharingClient:  sharingClient,
		reminderClient: reminderClient,
	}
}

func (h *handler) Todo(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}

	vars := mux.Vars(r)
	todoID := vars["todo_id"]

	todo, err := h.todoClient.GetTodoTask(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, todoID)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, todo)
}

func (h *handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	var req models.CreateTodoTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	todo, err := h.todoClient.CreateTodoTask(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, req)
	if err != nil {
		log.Printf("zakvo gurmish e batal %s", err.Error())
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, todo)
}

func (h *handler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	todoID := vars["todo_id"]

	var req models.UpdateTodoTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	todo, err := h.todoClient.UpdateTodoTask(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, todoID, req)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, todo)
}

func (h *handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	todoID := vars["todo_id"]

	err = h.todoClient.DeleteTodoTask(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, todoID)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	todos, err := h.todoClient.GetStandaloneTasks(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, todos)
}

func (h *handler) CreateTodoList(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	var req models.CreateTodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	list, err := h.todoClient.CreateTodoList(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, req)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, list)
}

func (h *handler) GetTodoLists(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	lists, err := h.todoClient.GetTodoListsWithTasks(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, lists)
}

func (h *handler) GetTodoList(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	listID := vars["list_id"]

	list, err := h.todoClient.GetTodoList(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, listID)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, list)
}

func (h *handler) UpdateTodoList(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	listID := vars["list_id"]

	var req models.UpdateTodoListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	list, err := h.todoClient.UpdateTodoList(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, listID, req)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, list)
}

func (h *handler) DeleteTodoList(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	listID := vars["list_id"]

	err = h.todoClient.DeleteTodoList(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, listID)
	if err != nil {
		if todoErr, ok := errors.AsType[clients.TodoServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, todoErr.Status, todoErr.ErrorMessage, todoErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Notes handlers

func (h *handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	notes, err := h.notesClient.GetAll(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID)
	if err != nil {
		if notesErr, ok := errors.AsType[clients.NotesServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, notesErr.Status, notesErr.ErrorMessage, notesErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, notes)
}

func (h *handler) GetNote(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	noteID := vars["note_id"]

	note, err := h.notesClient.GetByID(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, noteID)
	if err != nil {
		if notesErr, ok := errors.AsType[clients.NotesServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, notesErr.Status, notesErr.ErrorMessage, notesErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, note)
}

func (h *handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	var req models.NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	note, err := h.notesClient.Create(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, req)
	if err != nil {
		if notesErr, ok := errors.AsType[clients.NotesServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, notesErr.Status, notesErr.ErrorMessage, notesErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, note)
}

func (h *handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	noteID := vars["note_id"]

	var req models.NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	note, err := h.notesClient.Update(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, noteID, req)
	if err != nil {
		if notesErr, ok := errors.AsType[clients.NotesServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, notesErr.Status, notesErr.ErrorMessage, notesErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, note)
}

func (h *handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	noteID := vars["note_id"]

	err = h.notesClient.Delete(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, noteID)
	if err != nil {
		if notesErr, ok := errors.AsType[clients.NotesServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, notesErr.Status, notesErr.ErrorMessage, notesErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Sharing handlers

func (h *handler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	noteID := vars["note_id"]

	shareLink, err := h.sharingClient.CreateShareLink(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, noteID)
	if err != nil {
		if sharingErr, ok := errors.AsType[clients.SharingServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, sharingErr.Status, sharingErr.ErrorMessage, sharingErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, shareLink)
}

func (h *handler) OpenShareLink(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	vars := mux.Vars(r)
	token := vars["token"]

	sharedNote, err := h.sharingClient.OpenShareLink(ctx, token)
	if err != nil {
		if sharingErr, ok := errors.AsType[clients.SharingServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, sharingErr.Status, sharingErr.ErrorMessage, sharingErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, sharedNote)
}

// Reminder handlers

func (h *handler) GetReminders(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	filter := r.URL.Query().Get("status")

	reminders, err := h.reminderClient.GetReminders(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, filter)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, reminders)
}

func (h *handler) GetReminder(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	reminderID := vars["reminder_id"]

	reminder, err := h.reminderClient.GetReminderByID(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, reminderID)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, reminder)
}

func (h *handler) CreateReminder(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	var req models.ReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	reminder, err := h.reminderClient.CreateReminder(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, req)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusCreated, reminder)
}

func (h *handler) UpdateReminder(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	var req models.ReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return
	}

	reminder, err := h.reminderClient.UpdateReminder(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, req)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, reminder)
}

func (h *handler) DeleteReminder(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	reminderID := vars["reminder_id"]

	err = h.reminderClient.DeleteReminder(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, reminderID)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Notification handlers

func (h *handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	var readFilter *bool
	if readParam := r.URL.Query().Get("read"); readParam != "" {
		read := readParam == "true"
		readFilter = &read
	}

	notifications, err := h.reminderClient.GetNotifications(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, readFilter)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, notifications)
}

func (h *handler) GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	count, err := h.reminderClient.CountUnreadNotifications(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, count)
}

func (h *handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	vars := mux.Vars(r)
	notificationID := vars["notification_id"]

	notification, err := h.reminderClient.MarkNotificationAsRead(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID, notificationID)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	http_helpers.WriteSuccessResponse(w, http.StatusOK, notification)
}

func (h *handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	err = h.reminderClient.MarkAllNotificationsAsRead(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	contextWithUserMetadata, err := prepareContextAndExtractUserID(r)
	if err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, err.Error())
		return
	}
	defer contextWithUserMetadata.cancelFunc()

	err = h.reminderClient.DeleteAllNotifications(contextWithUserMetadata.ctx,
		contextWithUserMetadata.userID)
	if err != nil {
		if reminderErr, ok := errors.AsType[clients.ReminderServiceError](err); ok {
			http_helpers.WriteErrorResponse(w, reminderErr.Status, reminderErr.ErrorMessage, reminderErr.Message)
			return
		}

		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError, http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func prepareContextAndExtractUserID(r *http.Request) (*ctxWithUserMetadata, error) {
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		log.Printf("missing user id ama zashto")
		return nil, errors.New("missing userID in context")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

	return &ctxWithUserMetadata{
		ctx:        ctx,
		cancelFunc: cancel,
		userID:     userID.(string),
	}, nil
}
