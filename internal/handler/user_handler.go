package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"user-management-api/internal/middleware"
	"user-management-api/internal/model"
	"user-management-api/internal/service"
)

type UserHandler struct {
	svc      *service.UserService
	validate *validator.Validate
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc, validate: validator.New()}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "user ID must be a valid UUID"})
		return
	}

	u, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, u)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "message": "user ID must be a valid UUID"})
		return
	}

	if c.MustGet(middleware.UserIDKey).(uuid.UUID) != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden", "message": "you can only update your own profile"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": err.Error()})
		return
	}
	if req.Name == "" && req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": "at least one field (name, email) must be provided"})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error", "message": firstValidationError(err)})
		return
	}

	u, err := h.svc.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		fail(c, err)
		return
	}
	ok(c, u)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var q model.ListUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_query", "message": err.Error()})
		return
	}

	users, err := h.svc.ListUsers(c.Request.Context(), &q)
	if err != nil {
		fail(c, err)
		return
	}

	if users == nil {
		users = []*model.User{}
	}
	ok(c, users)
}
