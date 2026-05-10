package validate

import (
	"fmt"
	"net/mail"
	"strings"
)

func RequiredString(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("field '%s' is required and cannot be empty", field)
	}
	return nil
}

func StringMaxLen(field, value string, max int) error {
	if len(value) > max {
		return fmt.Errorf("field '%s' must be at most %d characters, got %d", field, max, len(value))
	}
	return nil
}

func PositiveInt(field string, value int) error {
	if value <= 0 {
		return fmt.Errorf("field '%s' must be > 0, got %d", field, value)
	}
	return nil
}

func Email(field, value string) error {
	if err := RequiredString(field, value); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("field '%s' must be a valid email address", field)
	}
	return nil
}
