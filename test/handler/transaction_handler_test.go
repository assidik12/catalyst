package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/handler"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/middleware"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionService mocks service.TransactionService interface.
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) Save(ctx context.Context, req dto.TransactionRequest, idUser int) (domain.Transaction, error) {
	args := m.Called(ctx, req, idUser)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	args := m.Called(ctx, idUser)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// ctxWithUserID injects a user ID into the request context, simulating auth middleware.
func ctxWithUserID(req *http.Request, userID int) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func TestGetAllTransactions_OK(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	userID := 1
	expectedTxs := []domain.Transaction{
		{ID: "tx-1", UserID: userID, TotalPrice: 10000, CreatedAt: time.Now()},
	}
	mockService.On("GetAll", mock.Anything, userID).Return(expectedTxs, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	req = ctxWithUserID(req, userID)
	rec := httptest.NewRecorder()

	txHandler.GetAllTransaction(rec, req, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestGetAllTransactions_MissingUserID(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	// No user ID injected in context
	req := httptest.NewRequest(http.MethodGet, "/transactions", nil)
	rec := httptest.NewRecorder()

	txHandler.GetAllTransaction(rec, req, nil)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertNotCalled(t, "GetAll")
}

func TestGetTransactionById_OK(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	txID := "tx-abc-123"
	expectedTx := domain.Transaction{ID: txID, UserID: 1, TotalPrice: 50000, CreatedAt: time.Now()}
	mockService.On("FindById", mock.Anything, txID).Return(expectedTx, nil)

	req := httptest.NewRequest(http.MethodGet, "/transactions/"+txID, nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: txID}}
	txHandler.GetTransactionById(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestGetTransactionById_NotFound(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	txID := "nonexistent"
	mockService.On("FindById", mock.Anything, txID).Return(domain.Transaction{}, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/transactions/"+txID, nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: txID}}
	txHandler.GetTransactionById(rec, req, params)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}

func TestCreateTransaction_OK(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	userID := 1
	reqBody := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ID: 1, Qty: 2}},
	}
	expectedTx := domain.Transaction{
		ID:         "tx-new-uuid",
		UserID:     userID,
		TotalPrice: 100000,
		CreatedAt:  time.Now(),
	}
	mockService.On("Save", mock.Anything, reqBody, userID).Return(expectedTx, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = ctxWithUserID(req, userID)
	rec := httptest.NewRecorder()

	txHandler.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockService.AssertExpectations(t)
}

func TestCreateTransaction_MissingUserID(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	reqBody := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ID: 1, Qty: 1}},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No user ID in context
	rec := httptest.NewRecorder()

	txHandler.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockService.AssertNotCalled(t, "Save")
}

func TestCreateTransaction_BadJSON(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req = ctxWithUserID(req, 1)
	rec := httptest.NewRecorder()

	txHandler.CreateTransaction(rec, req, nil)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockService.AssertNotCalled(t, "Save")
}

func TestDeleteTransaction_OK(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	txID := "tx-to-delete"
	mockService.On("Delete", mock.Anything, txID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/transactions/"+txID, nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: txID}}
	txHandler.DeleteTransaction(rec, req, params)

	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	mockService := new(MockTransactionService)
	txHandler := handler.NewTransactionHandler(mockService)

	txID := "ghost-tx"
	mockService.On("Delete", mock.Anything, txID).Return(domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/transactions/"+txID, nil)
	rec := httptest.NewRecorder()

	params := httprouter.Params{httprouter.Param{Key: "id", Value: txID}}
	txHandler.DeleteTransaction(rec, req, params)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockService.AssertExpectations(t)
}
