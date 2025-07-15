package main

import "fmt"

// ProjectNotFoundError represents an error when a project is not found
type ProjectNotFoundError struct {
	Permalink string
}

func (e *ProjectNotFoundError) Error() string {
	return fmt.Sprintf("project not found: %s", e.Permalink)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
}

// APIError represents an API error
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
