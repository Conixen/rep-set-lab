package validate_test

import (
	"testing"

	"github.com/leonj/rep-set-lab/internal/validate"
)

func TestRequiredString(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"valid", "hello", false},
		{"single char", "a", false},
		{"value with spaces", "hello world", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.RequiredString("field", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredString(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestStringMaxLen(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		max     int
		wantErr bool
	}{
		{"under limit", "abc", 5, false},
		{"at limit", "abcde", 5, false},
		{"over limit by one", "abcdef", 5, true},
		{"empty at limit 0", "", 0, false},
		{"one char over limit 0", "a", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.StringMaxLen("field", tt.value, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringMaxLen(%q, %d) error = %v, wantErr %v", tt.value, tt.max, err, tt.wantErr)
			}
		})
	}
}

func TestPositiveInt(t *testing.T) {
	tests := []struct {
		value   int
		wantErr bool
	}{
		{-10, true},
		{-1, true},
		{0, true},
		{1, false},
		{100, false},
	}
	for _, tt := range tests {
		err := validate.PositiveInt("field", tt.value)
		if (err != nil) != tt.wantErr {
			t.Errorf("PositiveInt(%d) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid standard", "user@example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"valid subdomain", "user@mail.example.com", false},
		{"empty", "", true},
		{"no at sign", "userexample.com", true},
		{"at only", "@", true},
		{"no domain", "user@", true},
		{"local part only", "user", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Email("email", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Email(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
