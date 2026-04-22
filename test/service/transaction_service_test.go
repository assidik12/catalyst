package service_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Save(ctx context.Context, user domain.User) (domain.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepo) FindById(ctx context.Context, id int) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

type MockTransactionRepo struct {
	mock.Mock
}

func (m *MockTransactionRepo) Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, tx, transaction)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	args := m.Called(ctx, idUser)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, topic string, data any) error {
	args := m.Called(ctx, topic, data)
	return args.Error(0)
}


func setupTransactionServiceTesting(t *testing.T) (*MockTransactionRepo, *MockProductRepo, *MockUserRepo, *MockProducer, service.TransactionService, sqlmock.Sqlmock) {
	mockTransactionRepo := new(MockTransactionRepo)
	mockProductRepo := new(MockProductRepo)
	mockUserRepo := new(MockUserRepo)
	mockProducer := new(MockProducer)

	db, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	validate := validator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	transactionService := service.NewTransactionService(
		mockTransactionRepo,
		db,
		validate,
		mockUserRepo,
		mockProductRepo,
		mockProducer,
		logger,
	)

	return mockTransactionRepo, mockProductRepo, mockUserRepo, mockProducer, transactionService, sqlMock
}

func TestSaveTransaction_ServerSidePriceCalculation(t *testing.T) {
	mockTransactionRepo, mockProductRepo, mockUserRepo, mockProducer, transactionService, sqlMock := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userId := 1

	// Setup valid user
	mockUserRepo.On("FindById", mock.Anything, userId).Return(domain.User{ID: userId, Email: "test@example.com"}, nil)

	// Setup product with server-side price of 50000.
	// We ignore whatever price might be inferred from the frontend, ensuring the backend drives pricing.
	mockProductRepo.On("FindById", mock.Anything, 101).Return(domain.Product{
		ID:    101,
		Name:  "Test Item",
		Price: 50000,
		Stock: 50,
	}, nil)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{
			{
				ID:  101,
				Qty: 2,
			},
		},
	}

	expectedTotalPrice := 100000 // 50000 * 2

	// Setup DB expectations for the transaction
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Assert that we are passing the calculated total price to the repository
	mockTransactionRepo.On("Save", mock.Anything, mock.AnythingOfType("*sql.Tx"), mock.MatchedBy(func(tx domain.Transaction) bool {
		return tx.TotalPrice == expectedTotalPrice
	})).Return(domain.Transaction{
		ID:         "tx-uuid-123",
		UserID:     userId,
		TotalPrice: expectedTotalPrice,
		CreatedAt:  time.Now(),
	}, nil)

	// Mock the producer correctly to avoid panic since it runs asynchronously
	mockProducer.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	tx, err := transactionService.Save(ctx, req, userId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotalPrice, tx.TotalPrice)

	mockTransactionRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestSaveTransaction_InsufficientStock(t *testing.T) {
	_, mockProductRepo, mockUserRepo, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userId := 1

	// Setup valid user
	mockUserRepo.On("FindById", mock.Anything, userId).Return(domain.User{ID: userId, Email: "test@example.com"}, nil)

	// Product has only 1 in stock
	mockProductRepo.On("FindById", mock.Anything, 101).Return(domain.Product{
		ID:    101,
		Name:  "Test Item",
		Price: 50000,
		Stock: 1,
	}, nil)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{
			{
				ID:  101,
				Qty: 5, // Requires 5, but stock is only 1
			},
		},
	}

	tx, err := transactionService.Save(ctx, req, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Contains(t, err.Error(), "insufficient stock")
	assert.Empty(t, tx) // Transaction should be zero-value struct

	mockProductRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestSaveTransaction_UserNotFound(t *testing.T) {
	_, _, mockUserRepo, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userId := 999

	mockUserRepo.On("FindById", mock.Anything, userId).Return(domain.User{}, domain.ErrNotFound)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ID: 101, Qty: 1}},
	}

	tx, err := transactionService.Save(ctx, req, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, tx)
	mockUserRepo.AssertExpectations(t)
}

func TestSaveTransaction_ProductNotFound(t *testing.T) {
	_, mockProductRepo, mockUserRepo, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userId := 1

	mockUserRepo.On("FindById", mock.Anything, userId).Return(domain.User{ID: userId, Email: "test@example.com"}, nil)
	mockProductRepo.On("FindById", mock.Anything, 999).Return(domain.Product{}, domain.ErrNotFound)

	req := dto.TransactionRequest{
		Products: []dto.TransactionItem{{ID: 999, Qty: 1}},
	}

	tx, err := transactionService.Save(ctx, req, userId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	assert.Empty(t, tx)
	mockProductRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestFindTransactionById_Success(t *testing.T) {
	mockTransactionRepo, _, _, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()
	txID := "tx-abc-123"

	expectedTx := domain.Transaction{
		ID:         txID,
		UserID:     1,
		TotalPrice: 50000,
	}

	mockTransactionRepo.On("FindById", mock.Anything, txID).Return(expectedTx, nil)

	tx, err := transactionService.FindById(ctx, txID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTx.ID, tx.ID)
	assert.Equal(t, expectedTx.TotalPrice, tx.TotalPrice)
	mockTransactionRepo.AssertExpectations(t)
}

func TestFindTransactionById_EmptyID(t *testing.T) {
	_, _, _, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()

	tx, err := transactionService.FindById(ctx, "")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Empty(t, tx)
}

func TestGetAllTransactions_Success(t *testing.T) {
	mockTransactionRepo, _, _, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()
	userID := 1

	expectedTxs := []domain.Transaction{
		{ID: "tx-1", UserID: userID, TotalPrice: 10000},
		{ID: "tx-2", UserID: userID, TotalPrice: 20000},
	}

	mockTransactionRepo.On("GetAll", mock.Anything, userID).Return(expectedTxs, nil)

	txs, err := transactionService.GetAll(ctx, userID)

	assert.NoError(t, err)
	assert.Len(t, txs, 2)
	mockTransactionRepo.AssertExpectations(t)
}

func TestDeleteTransaction_Success(t *testing.T) {
	mockTransactionRepo, _, _, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()
	txID := "tx-to-delete"

	mockTransactionRepo.On("Delete", mock.Anything, txID).Return(nil)

	err := transactionService.Delete(ctx, txID)

	assert.NoError(t, err)
	mockTransactionRepo.AssertExpectations(t)
}

func TestDeleteTransaction_EmptyID(t *testing.T) {
	_, _, _, _, transactionService, _ := setupTransactionServiceTesting(t)

	ctx := context.Background()

	err := transactionService.Delete(ctx, "")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}
