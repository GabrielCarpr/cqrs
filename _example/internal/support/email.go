package support

import (
	"example/internal/errs"
	"strings"
	"encoding/json"
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
	return json.Marshal(e.Email)
}

func (e *Email) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &e.Email)
}

func (e *Email) Bind(data interface{}) error {
	(*e).Email = data.(string)
	return nil
}
