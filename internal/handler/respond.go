package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"user-management-api/internal/repository"
	"user-management-api/internal/service"
)

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// fail maps a domain error to the appropriate HTTP status and error code.
// All error-to-HTTP mapping lives here — handlers stay free of switch/if chains.
func fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "user not found"})
	case errors.Is(err, repository.ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "conflict", "message": "email already in use"})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_credentials", "message": "email or password is incorrect"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

// firstValidationError returns a human-readable message for the first failed field.
func firstValidationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) && len(ve) > 0 {
		f := ve[0]
		switch f.Tag() {
		case "required":
			return f.Field() + " is required"
		case "email":
			return f.Field() + " must be a valid email address"
		case "min":
			return f.Field() + " must be at least " + f.Param() + " characters"
		}
		return f.Field() + " is invalid"
	}
	return err.Error()
}
