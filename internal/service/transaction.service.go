package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/event"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TransactionService defines the business-logic contract for transactions.
// (Previously mis-spelled TrancationService — fixed here.)
type TransactionService interface {
	Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error)
	FindById(ctx context.Context, id int) (domain.Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error)
	Delete(ctx context.Context, id int) error
}

type transactionService struct {
	repo      domain.TransactionRepository // ← domain interface, not mysql package
	userRepo  domain.UserRepository        // ← domain interface, not mysql package
	DB        *sql.DB
	validator *validator.Validate
	producer  event.Producer
}

// NewTransactionService constructs a TransactionService.
func NewTransactionService(
	repo domain.TransactionRepository,
	DB *sql.DB,
	validate *validator.Validate,
	userRepo domain.UserRepository,
	producer event.Producer,
) TransactionService {
	return &transactionService{
		repo:      repo,
		userRepo:  userRepo,
		DB:        DB,
		validator: validate,
		producer:  producer,
	}
}

// Delete implements TransactionService.
func (t *transactionService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("%w: transaction id must be positive", domain.ErrInvalidInput)
	}

	return t.repo.Delete(ctx, id)
}

// FindById implements TransactionService.
func (t *transactionService) FindById(ctx context.Context, id int) (domain.Transaction, error) {
	if id <= 0 {
		return domain.Transaction{}, fmt.Errorf("%w: transaction id must be positive", domain.ErrInvalidInput)
	}

	transaction, err := t.repo.FindById(ctx, id)
	if err != nil {
		return domain.Transaction{}, err
	}

	return transaction, nil
}

// GetAll implements TransactionService.
func (t *transactionService) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	return t.repo.GetAll(ctx, idUser)
}

// Save implements TransactionService.
func (t *transactionService) Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error) {
	// Verify the user actually exists before creating a transaction for them
	user, err := t.userRepo.FindById(ctx, idUser)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("%w: user id %d", domain.ErrNotFound, idUser)
	}

	transactionDetailID := uuid.NewString()

	// Map DTO products to domain.TransactionDetail using corrected field names
	products := make([]domain.TransactionDetail, 0, len(transaction.Products))
	for _, p := range transaction.Products {
		products = append(products, domain.TransactionDetail{
			ProductID: p.ID,
			Quantity:  p.Qty,
		})
	}

	transactionToSave := domain.Transaction{
		UserID:              user.ID,
		TransactionDetailID: transactionDetailID,
		TotalPrice:          transaction.TotalPrice,
		Products:            products,
	}

	tx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer tx.Rollback() //nolint:errcheck // rollback on deferred error is intentional

	savedTransaction, err := t.repo.Save(ctx, tx, transactionToSave)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Transaction{}, err
	}

	return savedTransaction, nil
}
