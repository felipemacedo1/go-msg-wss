package api

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidID         = errors.New("invalid ID format")
	ErrEmptyMessage      = errors.New("message cannot be empty")
	ErrMessageTooLong    = errors.New("message exceeds maximum length")
	ErrInvalidAuthorID   = errors.New("author ID cannot be empty")
	ErrInvalidAuthorName = errors.New("author name cannot be empty")
	ErrInvalidTheme      = errors.New("theme cannot be empty")
	ErrThemeTooLong      = errors.New("theme exceeds maximum length")
)

const (
	MaxMessageLength = 2000 // Maximum characters for a message
	MaxThemeLength   = 100  // Maximum characters for a room theme
)

// ValidateUUID validates if a string is a valid UUID
func ValidateUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidID
	}
	return nil
}

// ValidateMessage validates a message string
func ValidateMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return ErrEmptyMessage
	}

	if len(message) > MaxMessageLength {
		return ErrMessageTooLong
	}

	return nil
}

// ValidateAuthor validates author information
func ValidateAuthor(authorID, authorName string) error {
	if strings.TrimSpace(authorID) == "" {
		return ErrInvalidAuthorID
	}

	if strings.TrimSpace(authorName) == "" {
		return ErrInvalidAuthorName
	}

	return nil
}

// ValidateTheme validates a room theme
func ValidateTheme(theme string) error {
	if strings.TrimSpace(theme) == "" {
		return ErrInvalidTheme
	}

	if len(theme) > MaxThemeLength {
		return ErrThemeTooLong
	}

	return nil
}
