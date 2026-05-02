package response

const (
	ErrCodeInvalidRequestBody      = "INVALID_REQUEST_BODY"
	ErrCodeValidationFailed        = "VALIDATION_FAILED"
	ErrCodeTransactionBeginFailed  = "TRANSACTION_BEGIN_FAILED"
	ErrCodeEmailAlreadyExists      = "EMAIL_ALREADY_EXISTS"
	ErrCodeUserRegistrationFailed  = "USER_REGISTRATION_FAILED"
	ErrCodeTransactionCommitFailed = "TRANSACTION_COMMIT_FAILED"
	ErrCodeInternalServerError     = "INTERNAL_SERVER_ERROR"

	ErrCodeInvalidLoginCredentials = "INVALID_LOGIN_CREDENTIALS"
	ErrCodeUserLoginFailed         = "USER_LOGIN_FAILED"
	ErrCodeUserLogoutFailed        = "USER_LOGOUT_FAILED"
)
