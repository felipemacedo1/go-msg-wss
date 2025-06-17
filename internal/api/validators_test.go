package api

import (
	"testing"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		uuid    string
		wantErr bool
	}{
		{
			name:    "Valid UUID",
			uuid:    "123e4567-e89b-12d3-a456-426614174000",
			wantErr: false,
		},
		{
			name:    "Invalid UUID",
			uuid:    "not-a-uuid",
			wantErr: true,
		},
		{
			name:    "Empty UUID",
			uuid:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "Valid message",
			message: "Hello, World!",
			wantErr: false,
		},
		{
			name:    "Empty message",
			message: "",
			wantErr: true,
		},
		{
			name:    "Whitespace-only message",
			message: "   ",
			wantErr: true,
		},
		{
			name:    "Message too long",
			message: string(make([]rune, MaxMessageLength+1)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAuthor(t *testing.T) {
	tests := []struct {
		name       string
		authorID   string
		authorName string
		wantErr    bool
	}{
		{
			name:       "Valid author",
			authorID:   "user123",
			authorName: "John Doe",
			wantErr:    false,
		},
		{
			name:       "Empty author ID",
			authorID:   "",
			authorName: "John Doe",
			wantErr:    true,
		},
		{
			name:       "Empty author name",
			authorID:   "user123",
			authorName: "",
			wantErr:    true,
		},
		{
			name:       "Whitespace author ID",
			authorID:   "   ",
			authorName: "John Doe",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthor(tt.authorID, tt.authorName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTheme(t *testing.T) {
	tests := []struct {
		name    string
		theme   string
		wantErr bool
	}{
		{
			name:    "Valid theme",
			theme:   "Technical Discussion",
			wantErr: false,
		},
		{
			name:    "Empty theme",
			theme:   "",
			wantErr: true,
		},
		{
			name:    "Whitespace-only theme",
			theme:   "   ",
			wantErr: true,
		},
		{
			name:    "Theme too long",
			theme:   string(make([]rune, MaxThemeLength+1)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTheme(tt.theme)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTheme() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
