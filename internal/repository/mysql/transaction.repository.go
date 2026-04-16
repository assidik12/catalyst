package mysql

import (
	"context"
	"database/sql"

	"github.com/assidik12/go-restfull-api/internal/domain"
)

// transactionRepository is the MySQL implementation of domain.TransactionRepository.
type transactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository returns a domain.TransactionRepository backed by MySQL.
func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

// Save implements domain.TransactionRepository.
// It uses the provided *sql.Tx so the service layer controls commit/rollback.
func (t *transactionRepository) Save(ctx context.Context, tx *sql.Tx, transaction domain.Transaction) (domain.Transaction, error) {
	// 1. Insert into master table `transactions`
	qMaster := "INSERT INTO transactions (user_id, transaction_detail_id, total_price) VALUES (?, ?, ?)"
	result, err := tx.ExecContext(ctx, qMaster,
		transaction.UserID,
		transaction.TransactionDetailID,
		transaction.TotalPrice,
	)
	if err != nil {
		return domain.Transaction{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Transaction{}, err
	}
	transaction.ID = int(id)

	// 2. Insert each line-item into `transaction_details`
	qDetail := "INSERT INTO transaction_details (transaction_detail_id, quantity, product_id) VALUES (?, ?, ?)"
	for _, detail := range transaction.Products {
		_, err := tx.ExecContext(ctx, qDetail,
			transaction.TransactionDetailID,
			detail.Quantity,
			detail.ProductID,
		)
		if err != nil {
			return domain.Transaction{}, err
		}
	}

	return transaction, nil
}

// FindById implements domain.TransactionRepository.
func (t *transactionRepository) FindById(ctx context.Context, id int) (domain.Transaction, error) {
	var transaction domain.Transaction

	// 1. Fetch master record
	qMaster := "SELECT id, user_id, transaction_detail_id, total_price FROM transactions WHERE id = ?"
	err := t.db.QueryRowContext(ctx, qMaster, id).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.TransactionDetailID,
		&transaction.TotalPrice,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Transaction{}, domain.ErrNotFound
		}
		return domain.Transaction{}, err
	}

	// 2. Fetch detail rows
	qDetail := "SELECT quantity, product_id FROM transaction_details WHERE transaction_detail_id = ?"
	rows, err := t.db.QueryContext(ctx, qDetail, transaction.TransactionDetailID)
	if err != nil {
		return domain.Transaction{}, err
	}
	defer rows.Close()

	var details []domain.TransactionDetail
	for rows.Next() {
		var detail domain.TransactionDetail
		if err := rows.Scan(&detail.Quantity, &detail.ProductID); err != nil {
			return domain.Transaction{}, err
		}
		details = append(details, detail)
	}
	if err := rows.Err(); err != nil {
		return domain.Transaction{}, err
	}

	transaction.Products = details
	return transaction, nil
}

// GetAll implements domain.TransactionRepository.
func (t *transactionRepository) GetAll(ctx context.Context, idUser int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	// 1. Fetch all master records for this user
	qMaster := "SELECT id, user_id, transaction_detail_id, total_price FROM transactions WHERE user_id = ?"
	rowsMaster, err := t.db.QueryContext(ctx, qMaster, idUser)
	if err != nil {
		return nil, err
	}
	defer rowsMaster.Close()

	for rowsMaster.Next() {
		var transaction domain.Transaction
		err := rowsMaster.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.TransactionDetailID,
			&transaction.TotalPrice,
		)
		if err != nil {
			return nil, err
		}

		// 2. Fetch detail rows for each master record
		qDetail := "SELECT quantity, product_id FROM transaction_details WHERE transaction_detail_id = ?"
		rowsDetail, err := t.db.QueryContext(ctx, qDetail, transaction.TransactionDetailID)
		if err != nil {
			return nil, err
		}

		var details []domain.TransactionDetail
		for rowsDetail.Next() {
			var detail domain.TransactionDetail
			if err := rowsDetail.Scan(&detail.Quantity, &detail.ProductID); err != nil {
				rowsDetail.Close()
				return nil, err
			}
			details = append(details, detail)
		}
		if err := rowsDetail.Err(); err != nil {
			rowsDetail.Close()
			return nil, err
		}
		rowsDetail.Close()

		transaction.Products = details
		transactions = append(transactions, transaction)
	}

	if err := rowsMaster.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// Delete implements domain.TransactionRepository.
// It assumes ON DELETE CASCADE is set on the foreign key so that deleting from
// `transactions` automatically removes the associated `transaction_details` rows.
func (t *transactionRepository) Delete(ctx context.Context, id int) error {
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
