package models

type LoginCredentials struct {
	Email    string `json:"email" validate:"required,email_regex"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
