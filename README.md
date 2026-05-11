# Notes Cloud Auth Service

A Go microservice that provides authentication and authorization for the Notes Cloud platform.

## Overview

The auth-service handles:

1. **User Registration & Login**: Email/password authentication with email verification
2. **OAuth2/OIDC**: Third-party login via Google and GitLab
3. **JWT Token Management**: Access token generation and validation
4. **Refresh Token Rotation**: Secure token refresh with httpOnly cookies
5. **User Management**: Profile retrieval and identity linking

**Important:** This service is designed to run behind the API Gateway. All client requests should go through the gateway, not directly to this service.

## Authentication Flow

### Cookie-Based Token Strategy

The service uses a **hybrid cookie approach** for security:

- **Refresh Token** → httpOnly cookie (JavaScript cannot access, secure against XSS)
- **Access Token** → Regular cookie for OAuth (JavaScript can read for API calls)
- **Access Token** → JSON response for regular login (stored in localStorage by frontend)

### Regular Login Flow

**Endpoint:** `POST /authService/api/v1/login`

1. Client sends credentials to gateway → gateway proxies to auth-service
2. Auth-service validates credentials and generates TokenBundle
3. Auth-service sets `refresh_token` httpOnly cookie
4. Auth-service returns **only AccessToken** in JSON (not full TokenBundle)
5. Frontend stores access token in localStorage
6. Frontend includes access token in Authorization header for API calls
7. Frontend includes `credentials: 'include'` to send refresh token cookie

**Key Implementation:** `internal/auth/handler.go:48-100`

```go
http.SetCookie(w, &http.Cookie{
    Name:     "refresh_token",
    Value:    loginResponse.RefreshToken,
    MaxAge:   7 * 24 * 60 * 60, // 7 days
    HttpOnly: true,
    Secure:   false, // false for HTTP (local dev), true for production
    SameSite: http.SameSiteLaxMode,
    Path:     "/",
})
http_helpers.WriteSuccessResponse(w, http.StatusOK, loginResponse.AccessToken)
```

### OAuth Flow (Google/GitLab)

**Endpoints:**
- `GET /authService/api/v1/auth/google/start`
- `GET /authService/api/v1/auth/google/callback`
- `GET /authService/api/v1/auth/gitlab/start`
- `GET /authService/api/v1/auth/gitlab/callback`

1. User clicks "Continue with Google" → Browser navigates to `/auth/google/start` via gateway
2. Auth-service redirects to Google OAuth with state/nonce
3. User authenticates with Google
4. Google redirects back to `/auth/google/callback` via gateway
5. Auth-service exchanges code for tokens and validates ID token
6. Auth-service finds or creates user, linking OAuth identity
7. Auth-service generates TokenBundle
8. Auth-service sets **both cookies**:
   - `refresh_token` (httpOnly, 7 days)
   - `access_token` (readable by JS, 1 hour) — needed for frontend to read after redirect
9. Auth-service redirects to `FRONTEND_URL`
10. Frontend reads `access_token` cookie and saves session

**Key Implementation:** `internal/oidc/handler.go:106-209`

**Why OAuth sets both cookies?** Regular login is a fetch API call (can return JSON), but OAuth is a browser redirect flow. The frontend needs to read the access token after redirect, so it's set in a readable cookie. The refresh token is still httpOnly for security.

### Token Refresh Flow

**Endpoint:** `POST /authService/api/v1/refresh`

1. Client sends request with `credentials: 'include'` → refresh token sent in cookie
2. Auth-service validates refresh token from cookie
3. Auth-service generates new TokenBundle (rotates both tokens)
4. Auth-service sets new `refresh_token` cookie
5. Auth-service returns **only new AccessToken** in JSON

**Frontend auto-refresh:** The frontend's `fetchWithAuth()` automatically retries 401 responses after refreshing tokens.

**Key Implementation:** `internal/auth/handler.go:146-207`

### Logout Flow

**Endpoint:** `POST /authService/api/v1/logout`

1. Client POSTs to logout with `credentials: 'include'`
2. Auth-service reads refresh token from cookie
3. Auth-service revokes token in database
4. Auth-service clears refresh token cookie (MaxAge: -1)
5. Frontend clears localStorage tokens

**Key Implementation:** `internal/auth/handler.go:102-144`

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
| GET | `/authService/api/v1/users/{user_id}` | Get user by ID |

**Note:** All other endpoints (notes, todos, reminders, notifications, sharing) should be accessed through the API Gateway service, not directly through the auth service. The auth service focuses on authentication and user management only.

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

Dynamic OIDC provider discovery via `OIDC_<PROVIDER>_*` environment variables:

- `OIDC_GOOGLE_ENABLED` - Enable Google OAuth (default: `false`)
- `OIDC_GOOGLE_PROVIDER_TYPE` - Provider type (default: `google`)
- `OIDC_GOOGLE_ISSUER_URL` - OIDC issuer URL (`https://accounts.google.com`)
- `OIDC_GOOGLE_CLIENT_ID` - Google OAuth client ID
- `OIDC_GOOGLE_CLIENT_SECRET` - Google OAuth client secret
- `OIDC_GOOGLE_REDIRECT_URL` - OAuth callback URL (must go through gateway: `http://localhost:8090/api/v1/auth/google/callback`)
- `OIDC_GOOGLE_SCOPES` - OAuth scopes (comma-separated: `openid,email,profile`)

- `OIDC_GITLAB_ENABLED` - Enable GitLab OAuth (default: `false`)
- `OIDC_GITLAB_PROVIDER_TYPE` - Provider type (default: `gitlab`)
- `OIDC_GITLAB_ISSUER_URL` - OIDC issuer URL (`https://gitlab.com`)
- `OIDC_GITLAB_CLIENT_ID` - GitLab OAuth client ID
- `OIDC_GITLAB_CLIENT_SECRET` - GitLab OAuth client secret
- `OIDC_GITLAB_REDIRECT_URL` - OAuth callback URL (must go through gateway: `http://localhost:8090/api/v1/auth/gitlab/callback`)
- `OIDC_GITLAB_SCOPES` - OAuth scopes (comma-separated: `openid,email,profile`)

**Important:** OAuth redirect URLs must point to the API Gateway (port 8090), not directly to auth-service (port 8081). This ensures cookies are set from the gateway's domain/port for consistent cookie handling.

### Frontend

- `FRONTEND_URL` - Frontend URL to redirect after OAuth login (default: `http://localhost:5173`)

### Email (Resend)
- `RESEND_API_KEY` - Resend API key
- `RESEND_FROM_EMAIL` - From email address

**Note:** Backend service URLs (notes, todos, etc.) are no longer needed in auth-service since the API Gateway handles all proxying.

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

**Important:** Client applications should access the API Gateway (port 8090), not the auth-service directly. The auth-service runs on port 8081 internally within the cluster.

```bash
# Access via API Gateway (recommended)
kubectl port-forward -n notes-cloud svc/api-gateway 8090:8090

# Direct access to auth-service (for debugging only)
kubectl port-forward -n notes-cloud svc/auth-service 8081:8081
```

## System Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Client    │────▶│   API Gateway    │────▶│  Auth Service   │
│  (Browser)  │     │   (port 8090)    │     │   (port 8081)   │
│             │     │                  │     │                 │
│             │     │  - CORS          │     │ - Registration  │
│             │     │  - Auth Proxy    │     │ - Login/Logout  │
│             │     │  - JWT Valid.    │     │ - OAuth/OIDC    │
│             │     │  - Credentials   │     │ - Token Mgmt    │
│             │     │                  │     │ - User Mgmt     │
└─────────────┘     └──────────────────┘     └─────────────────┘
                             │
                             ├──────────────▶ Notes Service
                             ├──────────────▶ Todo Service
                             ├──────────────▶ Reminder Service
                             └──────────────▶ Sharing Service
```

**Key Points:**
- All client requests go through API Gateway
- Gateway handles CORS with `Access-Control-Allow-Credentials: true`
- Gateway proxies `/api/v1/auth/*` to auth-service
- Gateway validates JWT tokens for protected routes before proxying
- Auth-service sets cookies via gateway's domain/port for consistent cookie handling
