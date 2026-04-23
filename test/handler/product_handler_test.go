package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assidik12/catalyst/internal/delivery/http/handler"
	"github.com/assidik12/catalyst/internal/domain"
	"github.com/assidik12/catalyst/internal/delivery/http/dto"
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
