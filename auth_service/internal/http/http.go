package http_helpers

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	ErrCodeInvalidRequestBody      = "INVALID_REQUEST_BODY"
	ErrCodeValidationFailed        = "VALIDATION_FAILED"
	ErrCodeTransactionBeginFailed  = "TRANSACTION_BEGIN_FAILED"
	ErrCodeEmailAlreadyExists      = "EMAIL_ALREADY_EXISTS"
	ErrCodeUserRegistrationFailed  = "USER_REGISTRATION_FAILED"
	ErrCodeUserNotFound            = "USER_NOT_FOUND"
	ErrCodeTransactionCommitFailed = "TRANSACTION_COMMIT_FAILED"
	ErrCodeInternalServerError     = "INTERNAL_SERVER_ERROR"
	ErrCodeUnauthorized            = "UNAUTHORIZED"

	ErrCodeInvalidLoginCredentials = "INVALID_LOGIN_CREDENTIALS"
	ErrCodeUserLoginFailed         = "USER_LOGIN_FAILED"
	ErrCodeUserLogoutFailed        = "USER_LOGOUT_FAILED"
	ErrInvalidPasswordLength       = "INVALID_PASSWORD_LENGTH"

	ErrMissingTokenFromHeader = "MISSING_HEADER_TOKEN"
	ErrInvalidToken           = "INVALID_ACCESS_TOKEN"

	ErrInvalidSessionCookie          = "INVALID_SESSION_COOKIE"
	ErrInvalidOauthState             = "INVALID_OAUTH_STATE"
	ErrInvalidAuthCode               = "INVALID_AUTH_CODE"
	ErrFailedToLoginWithOIDCProvider = "FAILED_TO_LOGIN_WITH_OIDC_PROVIDER"
	ErrFailedToValidateIDToken       = "INVALID_ID_TOKEN"

	ErrUnauthorized            = "UNAUTHORIZED"
	ErrInvalidVerificationCode = "INVALID_VERIFICATION_CODE"
	ErrEmailNotVerified        = "EMAIL_NOT_VERIFIED"
	ErrRemoteServiceError      = "REMOTE_SERVICE_ERROR"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse[T any] struct {
	Data T `json:"data"`
}

func writeResponse(w http.ResponseWriter, statusCode int, payload any) {
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, code string, message string) {
	writeResponse(w, statusCode, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func WriteSuccessResponse[T any](w http.ResponseWriter, statusCode int, data T) {
	writeResponse(w, statusCode, SuccessResponse[T]{
		Data: data,
	})
}

func DecodeRequestBody[T any](w http.ResponseWriter, request *http.Request, v *T) error {
	if err := json.NewDecoder(request.Body).Decode(v); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, ErrCodeInvalidRequestBody, err.Error())
		return err
	}

	return nil
}
