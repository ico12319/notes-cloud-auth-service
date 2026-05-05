package request_models

type RegisterRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	Password *string `json:"password"`
	Name     string  `json:"name" validate:"required,min=2,max=256"`
}
