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
type TransactionService interface {
	Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error)
	FindById(ctx context.Context, id int) (domain.Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error)
	Delete(ctx context.Context, id int) error
}

type transactionService struct {
	repo        domain.TransactionRepository
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository // Added to fetch prices and check stock
	DB          *sql.DB
	validator   *validator.Validate
	producer    event.Producer
}

// NewTransactionService constructs a TransactionService.
func NewTransactionService(
	repo domain.TransactionRepository,
	DB *sql.DB,
	validate *validator.Validate,
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository, // Added dependency
	producer event.Producer,
) TransactionService {
	return &transactionService{
		repo:        repo,
		userRepo:    userRepo,
		productRepo: productRepo,
		DB:          DB,
		validator:   validate,
		producer:    producer,
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
	// 1. Verify user exists
	user, err := t.userRepo.FindById(ctx, idUser)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("%w: user id %d", domain.ErrNotFound, idUser)
	}

	// 2. Calculate TotalPrice and validate stock server-side
	var totalPrice int
	domainProducts := make([]domain.TransactionDetail, 0, len(transaction.Products))
	
	for _, item := range transaction.Products {
		product, err := t.productRepo.FindById(ctx, item.ID)
		if err != nil {
			return domain.Transaction{}, fmt.Errorf("%w: product id %d", domain.ErrNotFound, item.ID)
		}

		// Security: Check stock
		if product.Stock < item.Qty {
			return domain.Transaction{}, fmt.Errorf("%w: insufficient stock for product %d (available: %d, requested: %d)", 
				domain.ErrInvalidInput, item.ID, product.Stock, item.Qty)
		}

		// Security: Calculate price based on DB value, not client input
		totalPrice += product.Price * item.Qty

		domainProducts = append(domainProducts, domain.TransactionDetail{
			ProductID: item.ID,
			Quantity:  item.Qty,
		})
	}

	transactionDetailID := uuid.NewString()

	transactionToSave := domain.Transaction{
		UserID:              user.ID,
		TransactionDetailID: transactionDetailID,
		TotalPrice:          totalPrice, // Server-side calculated
		Products:            domainProducts,
	}

	tx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer tx.Rollback()

	savedTransaction, err := t.repo.Save(ctx, tx, transactionToSave)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Transaction{}, err
	}

	return savedTransaction, nil
}
