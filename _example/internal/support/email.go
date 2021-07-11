package support

import (
	"example/internal/errs"
	"strings"
)

// NewEmail creates an email value object
func NewEmail(email string) (Email, error) {
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return Email{}, errs.ValidationError("Email invalid")
	}
	return Email{email}, nil
}

// Email is an email address
type Email struct {
	Email string
}

func (e Email) String() string {
	return e.Email
}

func (e Email) MarshalJSON() ([]byte, error) {
	return []byte(e.String()), nil
}