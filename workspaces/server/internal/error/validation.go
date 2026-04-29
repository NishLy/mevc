package apperror

import (
	"fmt"

	"github.com/NishLy/go-fiber-boilerplate/internal/response"
	"github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err error) []response.ValidationError {
	var errs []response.ValidationError

	// Check if the error is actually from the validator package
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			errs = append(errs, response.ValidationError{
				Field: e.Field(), // e.g., "Email"
				Tag:   e.Tag(),   // e.g., "required"
				// Optional: You can create custom messages based on the tag
				Message: fmt.Sprintf("Field '%s' failed validation on tag '%s'", e.Field(), e.Tag()),
			})
		}
		return errs
	}

	// Fallback if it's a different type of error
	return nil
}
