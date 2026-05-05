package request_models

type RegisterRequest struct {
	Email     string  `json:"email" validate:"required,email"`
	Password  *string `json:"password"`
	FirstName string  `json:"firstName" validate:"required,min=2,max=256"`
	LastName  string  `json:"lastName" validate:"required,min=2,max=256"`
}
