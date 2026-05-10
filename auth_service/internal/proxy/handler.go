package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/clients"
	http_helpers "github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/http"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/middleware"
	"github.com/notes-in-the-cloud/notes-cloud-auth-service/internal/models"
)

const downstreamTimeout = 5 * time.Second

type handler struct {
	todoClient     *clients.TodoClient
	notesClient    *clients.NotesClient
	sharingClient  *clients.SharingClient
	reminderClient *clients.ReminderNotificationClient
}

func NewHandler(
	todoClient *clients.TodoClient,
	notesClient *clients.NotesClient,
	sharingClient *clients.SharingClient,
	reminderClient *clients.ReminderNotificationClient,
) *handler {
	return &handler{
		todoClient:     todoClient,
		notesClient:    notesClient,
		sharingClient:  sharingClient,
		reminderClient: reminderClient,
	}
}

type errorMapper func(err error) (status int, code, message string, ok bool)

func makeErrorMapper[E error](extract func(E) (int, string, string)) errorMapper {
	return func(err error) (int, string, string, bool) {
		if e, ok := errors.AsType[E](err); ok {
			status, code, msg := extract(e)
			return status, code, msg, true
		}
		return 0, "", "", false
	}
}

var (
	mapTodoError = makeErrorMapper[clients.TodoServiceError](
		func(e clients.TodoServiceError) (int, string, string) {
			return e.Status, http_helpers.ErrRemoteServiceError, e.Message
		})

	mapNotesError = makeErrorMapper[clients.NotesServiceError](
		func(e clients.NotesServiceError) (int, string, string) {
			return e.Status, e.ErrorMessage, e.Message
		})

	mapSharingError = makeErrorMapper[clients.SharingServiceError](
		func(e clients.SharingServiceError) (int, string, string) {
			return e.Status, e.ErrorMessage, e.Message
		})

	mapReminderError = makeErrorMapper[clients.ReminderServiceError](
		func(e clients.ReminderServiceError) (int, string, string) {
			return e.Status, e.ErrorMessage, e.Message
		})
)

func (h *handler) runAuthenticated(
	w http.ResponseWriter,
	r *http.Request,
	successStatus int,
	mapErr errorMapper,
	fn func(ctx context.Context, userID string) (any, error),
) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http_helpers.WriteErrorResponse(w, http.StatusUnauthorized,
			http_helpers.ErrUnauthorized, "missing userID in context")
		return
	}

	h.runPublic(w, r, successStatus, mapErr, func(ctx context.Context) (any, error) {
		return fn(ctx, userID)
	})
}

func (h *handler) runPublic(
	w http.ResponseWriter,
	r *http.Request,
	successStatus int,
	mapErr errorMapper,
	fn func(ctx context.Context) (any, error),
) {
	ctx, cancel := context.WithTimeout(r.Context(), downstreamTimeout)
	defer cancel()

	result, err := fn(ctx)
	if err != nil {
		log.Printf("downstream call failed: path=%s error=%v", r.URL.Path, err)
		if status, code, msg, ok := mapErr(err); ok {
			http_helpers.WriteErrorResponse(w, status, code, msg)
			return
		}
		http_helpers.WriteErrorResponse(w, http.StatusInternalServerError,
			http_helpers.ErrCodeInternalServerError, err.Error())
		return
	}

	if successStatus == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http_helpers.WriteSuccessResponse(w, successStatus, result)
}

func decodeBody[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http_helpers.WriteErrorResponse(w, http.StatusBadRequest,
			http_helpers.ErrCodeInvalidRequestBody, "invalid request body")
		return req, false
	}
	return req, true
}

func (h *handler) Todo(w http.ResponseWriter, r *http.Request) {
	todoID := mux.Vars(r)["todo_id"]
	h.runAuthenticated(w, r, http.StatusOK, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.GetTodoTask(ctx, userID, todoID)
		})
}

func (h *handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[models.CreateTodoTaskRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusCreated, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.CreateTodoTask(ctx, userID, req)
		})
}

func (h *handler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	todoID := mux.Vars(r)["todo_id"]
	req, ok := decodeBody[models.UpdateTodoTaskRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusOK, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.UpdateTodoTask(ctx, userID, todoID, req)
		})
}

func (h *handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	todoID := mux.Vars(r)["todo_id"]
	h.runAuthenticated(w, r, http.StatusNoContent, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return nil, h.todoClient.DeleteTodoTask(ctx, userID, todoID)
		})
}

func (h *handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	h.runAuthenticated(w, r, http.StatusOK, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.GetStandaloneTasks(ctx, userID)
		})
}

func (h *handler) CreateTodoList(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[models.CreateTodoListRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusCreated, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.CreateTodoList(ctx, userID, req)
		})
}

func (h *handler) GetTodoLists(w http.ResponseWriter, r *http.Request) {
	h.runAuthenticated(w, r, http.StatusOK, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.GetTodoListsWithTasks(ctx, userID)
		})
}

func (h *handler) GetTodoList(w http.ResponseWriter, r *http.Request) {
	listID := mux.Vars(r)["list_id"]
	h.runAuthenticated(w, r, http.StatusOK, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.GetTodoList(ctx, userID, listID)
		})
}

func (h *handler) UpdateTodoList(w http.ResponseWriter, r *http.Request) {
	listID := mux.Vars(r)["list_id"]
	req, ok := decodeBody[models.UpdateTodoListRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusOK, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return h.todoClient.UpdateTodoList(ctx, userID, listID, req)
		})
}

func (h *handler) DeleteTodoList(w http.ResponseWriter, r *http.Request) {
	listID := mux.Vars(r)["list_id"]
	h.runAuthenticated(w, r, http.StatusNoContent, mapTodoError,
		func(ctx context.Context, userID string) (any, error) {
			return nil, h.todoClient.DeleteTodoList(ctx, userID, listID)
		})
}

func (h *handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	h.runAuthenticated(w, r, http.StatusOK, mapNotesError,
		func(ctx context.Context, userID string) (any, error) {
			return h.notesClient.GetAll(ctx, userID)
		})
}

func (h *handler) GetNote(w http.ResponseWriter, r *http.Request) {
	noteID := mux.Vars(r)["note_id"]
	h.runAuthenticated(w, r, http.StatusOK, mapNotesError,
		func(ctx context.Context, userID string) (any, error) {
			return h.notesClient.GetByID(ctx, userID, noteID)
		})
}

func (h *handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[models.NoteRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusCreated, mapNotesError,
		func(ctx context.Context, userID string) (any, error) {
			return h.notesClient.Create(ctx, userID, req)
		})
}

func (h *handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	noteID := mux.Vars(r)["note_id"]
	req, ok := decodeBody[models.NoteRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusOK, mapNotesError,
		func(ctx context.Context, userID string) (any, error) {
			return h.notesClient.Update(ctx, userID, noteID, req)
		})
}

func (h *handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	noteID := mux.Vars(r)["note_id"]
	h.runAuthenticated(w, r, http.StatusNoContent, mapNotesError,
		func(ctx context.Context, userID string) (any, error) {
			return nil, h.notesClient.Delete(ctx, userID, noteID)
		})
}

func (h *handler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	noteID := mux.Vars(r)["note_id"]
	h.runAuthenticated(w, r, http.StatusCreated, mapSharingError,
		func(ctx context.Context, userID string) (any, error) {
			return h.sharingClient.CreateShareLink(ctx, userID, noteID)
		})
}

// OpenShareLink is public — no auth required, the token in the URL is the auth.
func (h *handler) OpenShareLink(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	h.runPublic(w, r, http.StatusOK, mapSharingError,
		func(ctx context.Context) (any, error) {
			return h.sharingClient.OpenShareLink(ctx, token)
		})
}

func (h *handler) GetReminders(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("status")
	h.runAuthenticated(w, r, http.StatusOK, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.GetReminders(ctx, userID, filter)
		})
}

func (h *handler) GetReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := mux.Vars(r)["reminder_id"]
	h.runAuthenticated(w, r, http.StatusOK, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.GetReminderByID(ctx, userID, reminderID)
		})
}

func (h *handler) CreateReminder(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[models.ReminderRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusCreated, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.CreateReminder(ctx, userID, req)
		})
}

func (h *handler) UpdateReminder(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeBody[models.ReminderRequest](w, r)
	if !ok {
		return
	}
	h.runAuthenticated(w, r, http.StatusOK, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.UpdateReminder(ctx, userID, req)
		})
}

func (h *handler) DeleteReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := mux.Vars(r)["reminder_id"]
	h.runAuthenticated(w, r, http.StatusNoContent, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return nil, h.reminderClient.DeleteReminder(ctx, userID, reminderID)
		})
}

func (h *handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	var readFilter *bool
	if readParam := r.URL.Query().Get("read"); readParam != "" {
		read := readParam == "true"
		readFilter = &read
	}
	h.runAuthenticated(w, r, http.StatusOK, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.GetNotifications(ctx, userID, readFilter)
		})
}

func (h *handler) GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	h.runAuthenticated(w, r, http.StatusOK, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.CountUnreadNotifications(ctx, userID)
		})
}

func (h *handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	notificationID := mux.Vars(r)["notification_id"]
	h.runAuthenticated(w, r, http.StatusOK, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return h.reminderClient.MarkNotificationAsRead(ctx, userID, notificationID)
		})
}

func (h *handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	h.runAuthenticated(w, r, http.StatusNoContent, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return nil, h.reminderClient.MarkAllNotificationsAsRead(ctx, userID)
		})
}

func (h *handler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	h.runAuthenticated(w, r, http.StatusNoContent, mapReminderError,
		func(ctx context.Context, userID string) (any, error) {
			return nil, h.reminderClient.DeleteAllNotifications(ctx, userID)
		})
}
