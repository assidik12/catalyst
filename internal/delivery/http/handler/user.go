package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/pkg/response"
	"github.com/assidik12/go-restfull-api/internal/service"
	"github.com/julienschmidt/httprouter"
)

// UserHandler handles HTTP requests for user endpoints.
type UserHandler struct {
	service service.UserService
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// handleServiceError maps domain sentinel errors to the appropriate HTTP response.
func (h *UserHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.NotFound(w, err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		response.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		response.Unauthorized(w, err.Error())
	case errors.Is(err, domain.ErrConflict):
		response.Conflict(w, err.Error())
	default:
		response.InternalServerError(w, "internal server error")
	}
}

// Register handles POST /api/v1/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.Created(w, user)
}

// Login handles POST /api/v1/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	token, err := h.service.Login(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	response.OK(w, token)
}
