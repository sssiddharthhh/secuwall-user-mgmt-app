package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"user-management-api/internal/model"
	"user-management-api/internal/service"
)

type AuthHandler struct {
	svc      *service.UserService
	validate *validator.Validate
}

func NewAuthHandler(svc *service.UserService) *AuthHandler {
	return &AuthHandler{svc: svc, validate: validator.New()}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": firstValidationError(err)})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		fail(c, err)
		return
	}
	created(c, resp)
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req model.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": firstValidationError(err)})
		return
	}

	resp, err := h.svc.SignIn(c.Request.Context(), &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, resp)
}
