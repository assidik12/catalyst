package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/assidik12/go-restfull-api/internal/delivery/http/dto"
	"github.com/assidik12/go-restfull-api/internal/domain"
	"github.com/assidik12/go-restfull-api/internal/event"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TransactionService defines the business-logic contract for transactions.
type TransactionService interface {
	Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error)
	FindById(ctx context.Context, id string) (domain.Transaction, error)
	GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error)
	Delete(ctx context.Context, id string) error
}

type transactionService struct {
	repo        domain.TransactionRepository
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	DB          *sql.DB
	validator   *validator.Validate
	producer    event.Producer
	logger      *slog.Logger
}

func NewTransactionService(
	repo domain.TransactionRepository,
	DB *sql.DB,
	validate *validator.Validate,
	userRepo domain.UserRepository,
	productRepo domain.ProductRepository,
	producer event.Producer,
	logger *slog.Logger,
) TransactionService {
	return &transactionService{
		repo:        repo,
		userRepo:    userRepo,
		productRepo: productRepo,
		DB:          DB,
		validator:   validate,
		producer:    producer,
		logger:      logger,
	}
}

func (t *transactionService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: transaction id must not be empty", domain.ErrInvalidInput)
	}

	return t.repo.Delete(ctx, id)
}

func (t *transactionService) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	if id == "" {
		return domain.Transaction{}, fmt.Errorf("%w: transaction id must not be empty", domain.ErrInvalidInput)
	}

	transaction, err := t.repo.FindById(ctx, id)
	if err != nil {
		return domain.Transaction{}, err
	}

	return transaction, nil
}

func (t *transactionService) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	return t.repo.GetAll(ctx, idUser)
}

func (t *transactionService) Save(ctx context.Context, transaction dto.TransactionRequest, idUser int) (domain.Transaction, error) {
	// 1. Verify user exists
	user, err := t.userRepo.FindById(ctx, idUser)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("%w: user id %d", domain.ErrNotFound, idUser)
	}

	// 2. Start database transaction BEFORE the loop
	tx, err := t.DB.BeginTx(ctx, nil)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer tx.Rollback()

	// 3. Calculate TotalPrice, validate stock, decrement stock, and populate details
	var totalPrice int
	domainProducts := make([]domain.TransactionDetail, 0, len(transaction.Products))
	eventProducts := make([]event.ProductItem, 0, len(transaction.Products))

	for _, item := range transaction.Products {
		product, err := t.productRepo.FindById(ctx, item.ProductID)
		if err != nil {
			return domain.Transaction{}, fmt.Errorf("%w: product id %d", domain.ErrNotFound, item.ProductID)
		}

		if product.Stock < item.Quantity {
			return domain.Transaction{}, fmt.Errorf("%w: insufficient stock for product %d", domain.ErrInvalidInput, item.ProductID)
		}

		if err := t.productRepo.DecrementStock(ctx, tx, item.ProductID, item.Quantity); err != nil {
			return domain.Transaction{}, fmt.Errorf("failed to decrement stock: %w", err)
		}

		totalPrice += product.Price * item.Quantity

		domainProducts = append(domainProducts, domain.TransactionDetail{
			ProductID: item.ProductID,
			Price:     product.Price, // Capture price at time of transaction
			Quantity:  item.Quantity,
		})

		eventProducts = append(eventProducts, event.ProductItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	transactionToSave := domain.Transaction{
		ID:         uuid.NewString(), // Use UUID as primary key string
		UserID:     user.ID,
		TotalPrice: totalPrice,
		CreatedAt:  time.Now(),
		Products:   domainProducts,
	}

	savedTransaction, err := t.repo.Save(ctx, tx, transactionToSave)
	if err != nil {
		return domain.Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Transaction{}, err
	}

	// 3. Publish Event asynchronously
	orderEvent := event.OrderCreatedEvent{
		TransactionID: savedTransaction.ID,
		UserID:        savedTransaction.UserID,
		TotalPrice:    savedTransaction.TotalPrice,
		Products:      eventProducts,
		CreatedAt:     savedTransaction.CreatedAt,
	}

	go func() {
		pubCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := t.producer.Publish(pubCtx, event.TopicOrderCreated, orderEvent); err != nil {
			t.logger.Error("failed to publish order event",
				"transaction_id", orderEvent.TransactionID,
				"error", err,
			)
		} else {
			t.logger.Info("successfully published order event",
				"transaction_id", orderEvent.TransactionID,
			)
		}
	}()

	return savedTransaction, nil
}
