package auth

type Register struct {
	Username string `json:"username" validate:"required,min=6"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
