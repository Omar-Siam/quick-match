package validation

import (
	"github.com/go-playground/validator/v10"
	"quick-match/internal/models"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^.+@\S+\.\S+$`)

func emailRegexValidation(fl validator.FieldLevel) bool {
	return emailRegex.MatchString(fl.Field().String())
}

// ValidateLogin uses the validator package to validate the LoginCredentials struct,
// including a custom regex validation for the email.
func ValidateLogin(login models.LoginCredentials) error {
	validate := validator.New()
	validate.RegisterValidation("email_regex", emailRegexValidation)

	return validate.Struct(login)
}
