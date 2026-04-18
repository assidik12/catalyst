package mysql

import (
	"context"
	"database/sql"

	"github.com/assidik12/go-restfull-api/internal/domain"
)

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (t *transactionRepository) Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error) {
	// 1. Insert into master table `transactions`
	qMaster := "INSERT INTO transactions (id, user_id, total_price, created_at) VALUES (?, ?, ?, ?)"
	_, err := tx.ExecContext(ctx, qMaster,
		transaction.ID,
		transaction.UserID,
		transaction.TotalPrice,
		transaction.CreatedAt,
	)
	if err != nil {
		return domain.Transaction{}, err
	}

	// 2. Insert each line-item into `transaction_details`
	qDetail := "INSERT INTO transaction_details (transaction_id, product_id, price, quantity) VALUES (?, ?, ?, ?)"
	for _, detail := range transaction.Products {
		_, err := tx.ExecContext(ctx, qDetail,
			transaction.ID,
			detail.ProductID,
			detail.Price,
			detail.Quantity,
		)
		if err != nil {
			return domain.Transaction{}, err
		}
	}

	return transaction, nil
}

func (t *transactionRepository) FindById(ctx context.Context, id string) (domain.Transaction, error) {
	var transaction domain.Transaction

	qMaster := "SELECT id, user_id, total_price, created_at FROM transactions WHERE id = ?"
	err := t.db.QueryRowContext(ctx, qMaster, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.TotalPrice,
		&transaction.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Transaction{}, domain.ErrNotFound
		}
		return domain.Transaction{}, err
	}

	qDetail := "SELECT product_id, price, quantity FROM transaction_details WHERE transaction_id = ?"
	rows, err := t.db.QueryContext(ctx, qDetail, transaction.ID)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer rows.Close()

	var details []domain.TransactionDetail
	for rows.Next() {
		var detail domain.TransactionDetail
		if err := rows.Scan(&detail.ProductID, &detail.Price, &detail.Quantity); err != nil {
			return domain.Transaction{}, err
		}
		details = append(details, detail)
	}

	transaction.Products = details
	return transaction, nil
}

func (t *transactionRepository) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	qMaster := "SELECT id, user_id, total_price, created_at FROM transactions WHERE user_id = ?"
	rowsMaster, err := t.db.QueryContext(ctx, qMaster, idUser)
	if err != nil {
		return nil, err
	}
	defer rowsMaster.Close()

	var transactions []domain.Transaction
	for rowsMaster.Next() {
		var transaction domain.Transaction
		err := rowsMaster.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.TotalPrice,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		qDetail := "SELECT product_id, price, quantity FROM transaction_details WHERE transaction_id = ?"
		rowsDetail, err := t.db.QueryContext(ctx, qDetail, transaction.ID)
		if err != nil {
			return nil, err
		}

		var details []domain.TransactionDetail
		for rowsDetail.Next() {
			var detail domain.TransactionDetail
			if err := rowsDetail.Scan(&detail.ProductID, &detail.Price, &detail.Quantity); err != nil {
				rowsDetail.Close()
				return nil, err
			}
			details = append(details, detail)
		}
		rowsDetail.Close()

		transaction.Products = details
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (t *transactionRepository) Delete(ctx context.Context, id string) error {
	q := "DELETE FROM transactions WHERE id = ?"
	result, err := t.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
