package main

import "strings"

// Validator handles input validation
type Validator struct{}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateProjectPermalink validates a project permalink
func (v *Validator) ValidateProjectPermalink(permalink string) error {
	if strings.TrimSpace(permalink) == "" {
		return &ValidationError{
			Field:   "project_permalink",
			Message: "cannot be empty",
		}
	}
	return nil
}

// ValidateToolArguments validates tool arguments and extracts project permalink
func (v *Validator) ValidateToolArguments(arguments map[string]interface{}) (string, error) {
	projectPermalink, ok := arguments["project_permalink"].(string)
	if !ok {
		return "", &ValidationError{
			Field:   "project_permalink",
			Message: "must be a string",
		}
	}

	return projectPermalink, v.ValidateProjectPermalink(projectPermalink)
}
