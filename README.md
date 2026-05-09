# Notes Cloud Auth Service

A Go microservice that provides authentication, authorization, and acts as an API gateway/proxy for the Notes Cloud platform.

## Overview

The auth-service serves two primary functions:

1. **Authentication**: Handles user registration, login, JWT token management, and OAuth2/OIDC flows (Google, GitLab)
2. **API Gateway/Proxy**: Secures all backend microservices by validating JWT tokens, extracting user identity, and forwarding requests with user context

### How the Proxy Works

All protected endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

The service:
1. Extracts the JWT token from the `Authorization` header
2. Validates the token signature and expiration
3. Extracts the `user_id` from the token claims
4. Stores the `user_id` in the request context
5. Forwards the request to the appropriate backend service with the user identity

This centralizes authentication logic and allows backend services to focus on business logic without handling auth concerns.

## API Endpoints

### Health & Readiness

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/healthz` | Health check |
| GET | `/authService/api/v1/readyz` | Readiness check |

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/authService/api/v1/register` | Register a new user |
| POST | `/authService/api/v1/login` | Login with email/password |
| POST | `/authService/api/v1/logout` | Logout (invalidate refresh token) |
| POST | `/authService/api/v1/refresh` | Refresh access token |

### Email Verification (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/authService/api/v1/email/verify` | Verify email with verification code |
| POST | `/authService/api/v1/email/resend-verification` | Resend verification email |

### OAuth2/OIDC (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/auth/google/start` | Start Google OAuth flow |
| GET | `/authService/api/v1/auth/google/callback` | Google OAuth callback |
| GET | `/authService/api/v1/auth/gitlab/start` | Start GitLab OAuth flow |
| GET | `/authService/api/v1/auth/gitlab/callback` | GitLab OAuth callback |

### User (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/me` | Get current user info |

### Notes (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/notes` | Get all notes for user |
| POST | `/authService/api/v1/notes` | Create a new note |
| GET | `/authService/api/v1/notes/{note_id}` | Get a specific note |
| PUT | `/authService/api/v1/notes/{note_id}` | Update a note |
| DELETE | `/authService/api/v1/notes/{note_id}` | Delete a note |

### Sharing

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/authService/api/v1/notes/{note_id}/share-links` | Protected | Create a share link for a note |
| GET | `/authService/api/v1/share-links/{token}` | Public | Open a shared note via token |

### Todo Tasks (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/todos` | Get all standalone tasks |
| POST | `/authService/api/v1/todos` | Create a new task |
| GET | `/authService/api/v1/todos/{todo_id}` | Get a specific task |
| PUT | `/authService/api/v1/todos/{todo_id}` | Update a task |
| DELETE | `/authService/api/v1/todos/{todo_id}` | Delete a task |

### Todo Lists (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/todo-lists` | Get all todo lists with tasks |
| POST | `/authService/api/v1/todo-lists` | Create a new todo list |
| GET | `/authService/api/v1/todo-lists/{list_id}` | Get a specific todo list |
| PUT | `/authService/api/v1/todo-lists/{list_id}` | Update a todo list |
| DELETE | `/authService/api/v1/todo-lists/{list_id}` | Delete a todo list |

### Reminders (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/reminders` | Get all reminders (supports `?status=PENDING\|COMPLETED`) |
| POST | `/authService/api/v1/reminders` | Create a new reminder |
| PUT | `/authService/api/v1/reminders` | Update a reminder |
| GET | `/authService/api/v1/reminders/{reminder_id}` | Get a specific reminder |
| DELETE | `/authService/api/v1/reminders/{reminder_id}` | Delete a reminder |

### Notifications (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authService/api/v1/notifications` | Get all notifications (supports `?read=true\|false`) |
| DELETE | `/authService/api/v1/notifications` | Delete all notifications |
| GET | `/authService/api/v1/notifications/unread-count` | Get unread notification count |
| POST | `/authService/api/v1/notifications/read-all` | Mark all notifications as read |
| POST | `/authService/api/v1/notifications/{notification_id}/read` | Mark a notification as read |

## Configuration

The service is configured via environment variables:

### Database
- `DB_HOST` - PostgreSQL host (default: `localhost`)
- `DB_PORT` - PostgreSQL port (default: `5432`)
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `DB_SSLMODE` - SSL mode (default: `disable`)

### JWT & Tokens
- `JWT_SECRET` - Secret for signing access tokens
- `JWT_ISSUER` - Token issuer
- `JWT_AUDIENCE` - Token audience
- `JWT_TTL` - Access token TTL (e.g., `15m`)
- `REFRESH_TOKEN_SECRET` - Secret for refresh tokens
- `COOKIE_SECRET` - Secret for OAuth state cookies

### OAuth Providers
- `OIDC_GOOGLE_CLIENT_ID` / `OIDC_GOOGLE_CLIENT_SECRET`
- `OIDC_GITLAB_CLIENT_ID` / `OIDC_GITLAB_CLIENT_SECRET`

### Email (Resend)
- `RESEND_API_KEY` - Resend API key
- `RESEND_FROM_EMAIL` - From email address

### Backend Services
- `NOTES_SERVICE_URL` - Notes service URL (e.g., `http://notes-service:8082`)
- `TODO_SERVICE_URL` - Todo service URL (e.g., `http://todo-service:8085`)
- `REMINDER_SERVICE_URL` - Reminder service URL (e.g., `http://reminder-service:8084`)
- `SHARING_SERVICE_URL` - Sharing service URL (e.g., `http://sharing-service:8083`)

## Running Locally

```bash
# Set required environment variables
export DB_HOST=localhost
export DB_PASSWORD=your_password
export JWT_SECRET=your_jwt_secret
# ... other required vars

# Run the service
go run cmd/auth_service/main.go
```

The service starts on `http://localhost:8081`.

## Kubernetes Deployment

The service is deployed as part of the notes-cloud platform. See the `notes-cloud-infrastructure` repository for Kubernetes manifests.

```bash
# Port-forward to access locally
kubectl port-forward -n notes-cloud svc/auth-service 8081:8081
```

## Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Client    │────▶│   Auth Service   │────▶│  Notes Service  │
│             │     │                  │     └─────────────────┘
│             │     │  - Auth/JWT      │     ┌─────────────────┐
│             │     │  - Token Valid.  │────▶│  Todo Service   │
│             │     │  - User Context  │     └─────────────────┘
│             │     │  - Proxy/Gateway │     ┌─────────────────┐
│             │     │                  │────▶│ Reminder Service│
└─────────────┘     └──────────────────┘     └─────────────────┘
                                             ┌─────────────────┐
                                        ────▶│ Sharing Service │
                                             └─────────────────┘
```
