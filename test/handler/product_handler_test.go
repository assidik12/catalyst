package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/handler"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProductService matches service.ProductService interface
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) GetAllProducts(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	args := m.Called(ctx, page, pageSize)
	return args.Get(0).([]domain.Product), args.Error(1)
}

func (m *MockProductService) GetProductById(ctx context.Context, id int) (domain.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductService) CreateProduct(ctx context.Context, product dto.ProductRequest) (domain.Product, error) {
	args := m.Called(ctx, product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, id int, product dto.ProductRequest) (domain.Product, error) {
	args := m.Called(ctx, id, product)
	return args.Get(0).(domain.Product), args.Error(1)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestGetProductById_OK(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	expectedProduct := domain.Product{
		ID:   1,
		Name: "Test Target Product",
	}
	mockService.On("GetProductById", mock.Anything, 1).Return(expectedProduct, nil)

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "1"}}
	productHandler.GetProductById(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestGetProductById_NotFound(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	mockService.On("GetProductById", mock.Anything, 99).Return(domain.Product{}, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/products/99", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "99"}}
	productHandler.GetProductById(rec, req, params)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}

func TestContextCancellation_HandlingAbortAppropriately(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	// Simulate context cancellation internally
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // instantly cancel the context so the handler perceives an aborted request

	expectedErr := context.Canceled
	mockService.On("GetProductById", mock.Anything, 1).Return(domain.Product{}, expectedErr)

	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "1"}}
	productHandler.GetProductById(rec, req, params)

	// The product handler default handler uses handleServiceError which cascades to 500
	// for generic fallback errors like context.Canceled if it hasn't mapped it strictly.
	// We are verifying that the service mock recognized the aborted context.
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	
	// To actually verify Context cancellation was passed to the service layer:
	mockService.AssertCalled(t, "GetProductById", mock.MatchedBy(func(ctx context.Context) bool {
		return errors.Is(ctx.Err(), context.Canceled)
	}), 1)
}

func TestGetProductById_InvalidID(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/products/abc", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "abc"}}
	productHandler.GetProductById(rec, req, params)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "GetProductById")
}

func TestGetAllProducts_OK(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	expectedProducts := []domain.Product{
		{ID: 1, Name: "Product A"},
		{ID: 2, Name: "Product B"},
	}
	mockService.On("GetAllProducts", mock.Anything, 1, 10).Return(expectedProducts, nil)

	req := httptest.NewRequest(http.MethodGet, "/products?page=1&pageSize=10", nil)
	rec := httptest.NewRecorder()

	productHandler.GetAllProducts(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestGetAllProducts_DefaultPagination(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	mockService.On("GetAllProducts", mock.Anything, 1, 10).Return([]domain.Product{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rec := httptest.NewRecorder()

	productHandler.GetAllProducts(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestGetAllProducts_InvalidPage(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/products?page=0", nil)
	rec := httptest.NewRecorder()

	productHandler.GetAllProducts(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "GetAllProducts")
}

func TestGetAllProducts_InvalidPageSize(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/products?page=1&pageSize=abc", nil)
	rec := httptest.NewRecorder()

	productHandler.GetAllProducts(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "GetAllProducts")
}

func TestCreateProduct_OK(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	reqBody := dto.ProductRequest{
		Name:        "New Product",
		Price:       15000,
		Stock:       10,
		Img:         "img.png",
		Description: "Description",
		CategoryId:  1,
	}
	createdProduct := domain.Product{ID: 5, Name: reqBody.Name, Price: reqBody.Price}
	mockService.On("CreateProduct", mock.Anything, reqBody).Return(createdProduct, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	productHandler.CreateProduct(rec, req, nil)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockService.AssertExpectations(t)
}

func TestCreateProduct_BadJSON(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	productHandler.CreateProduct(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "CreateProduct")
}

func TestCreateProduct_ServiceError(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	reqBody := dto.ProductRequest{Name: "Bad"}
	mockService.On("CreateProduct", mock.Anything, reqBody).Return(domain.Product{}, domain.ErrInvalidInput)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	productHandler.CreateProduct(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateProduct_OK(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	reqBody := dto.ProductRequest{
		Name:        "Updated Product",
		Price:       25000,
		Stock:       15,
		Img:         "updated.png",
		Description: "Updated desc",
		CategoryId:  2,
	}
	updatedProduct := domain.Product{ID: 1, Name: reqBody.Name}
	mockService.On("UpdateProduct", mock.Anything, 1, reqBody).Return(updatedProduct, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "1"}}
	productHandler.UpdateProduct(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateProduct_InvalidID(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	req := httptest.NewRequest(http.MethodPut, "/products/xyz", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "xyz"}}
	productHandler.UpdateProduct(rec, req, params)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "UpdateProduct")
}

func TestDeleteProduct_OK(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	mockService.On("DeleteProduct", mock.Anything, 1).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "1"}}
	productHandler.DeleteProduct(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteProduct_InvalidID(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	req := httptest.NewRequest(http.MethodDelete, "/products/xyz", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "xyz"}}
	productHandler.DeleteProduct(rec, req, params)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "DeleteProduct")
}

func TestDeleteProduct_NotFound(t *testing.T) {
	mockService := new(MockProductService)
	productHandler := handler.NewProductHandler(mockService)

	mockService.On("DeleteProduct", mock.Anything, 999).Return(domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/products/999", nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: "999"}}
	productHandler.DeleteProduct(rec, req, params)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}
