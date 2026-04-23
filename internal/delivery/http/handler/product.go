package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/assidik12/catalyst/internal/delivery/http/dto"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/pkg/response"
	"github.com/assidik12/catalyst/internal/service"
	"github.com/julienschmidt/httprouter"
)

// ProductHandler handles HTTP requests for product endpoints.
type ProductHandler struct {
	service service.ProductService
}

// NewProductHandler constructs a ProductHandler.
func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// handleServiceError maps domain sentinel errors to the appropriate HTTP response.
// It is a shared helper used by every handler method in this file.
func (ph *ProductHandler) handleServiceError(w http.ResponseWriter, err error) {
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

// GetAllProducts handles GET /api/v1/products?page=1&pageSize=10
func (ph *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		response.BadRequest(w, "invalid page parameter")
		return
	}

	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		response.BadRequest(w, "invalid pageSize parameter")
		return
	}

	products, err := ph.service.GetAllProducts(r.Context(), page, pageSize)
	if err != nil {
		ph.handleServiceError(w, err)
		return
	}

	response.OK(w, products)
}

// GetProductById handles GET /api/v1/products/:id
func (ph *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idInt, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		response.BadRequest(w, "invalid product id")
		return
	}

	product, err := ph.service.GetProductById(r.Context(), idInt)
	if err != nil {
		ph.handleServiceError(w, err)
		return
	}

	response.OK(w, product)
}

// CreateProduct handles POST /api/v1/products
func (ph *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	product, err := ph.service.CreateProduct(r.Context(), req)
	if err != nil {
		ph.handleServiceError(w, err)
		return
	}

	response.Created(w, product)
}

// UpdateProduct handles PUT /api/v1/products/:id
func (ph *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idInt, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		response.BadRequest(w, "invalid product id")
		return
	}

	var req dto.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	product, err := ph.service.UpdateProduct(r.Context(), idInt, req)
	if err != nil {
		ph.handleServiceError(w, err)
		return
	}

	response.OK(w, product)
}

// DeleteProduct handles DELETE /api/v1/products/:id
func (ph *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idInt, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		response.BadRequest(w, "invalid product id")
		return
	}

	if err := ph.service.DeleteProduct(r.Context(), idInt); err != nil {
		ph.handleServiceError(w, err)
		return
	}

	response.OK(w, nil)
}
