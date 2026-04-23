package mysql

import (
	"context"
	"database/sql"

	"github.com/assidik12/go-restfull-api/internal/domain"
)

// productRepository is the MySQL implementation of domain.ProductRepository.
type productRepository struct {
	db *sql.DB
}

// NewProductRepository returns a domain.ProductRepository backed by MySQL.
func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

// Delete implements domain.ProductRepository.
func (p *productRepository) Delete(ctx context.Context, id int) error {
	q := "DELETE FROM products WHERE id = ?"

	_, err := p.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	return nil
}

// FindById implements domain.ProductRepository using QueryRowContext for efficiency.
func (p *productRepository) FindById(ctx context.Context, id int) (domain.Product, error) {
	// Explicit column names prevent breakages from schema changes
	q := "SELECT id, name, price, stock, description, img, category_id FROM products WHERE id = ?"

	var product domain.Product
	err := p.db.QueryRowContext(ctx, q, id).Scan(
		&product.ID,
		&product.Name,
		&product.Price,
		&product.Stock,
		&product.Description,
		&product.Img,
		&product.CategoryID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, err
	}

	return product, nil
}

// GetAll implements domain.ProductRepository with explicit columns and deferred Close.
func (p *productRepository) GetAll(ctx context.Context, page int, pageSize int) ([]domain.Product, error) {
	offset := (page - 1) * pageSize
	query := "SELECT id, name, price, stock, description, img, category_id FROM products LIMIT ? OFFSET ?"

	rows, err := p.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	// Defer immediately to prevent connection leaks
	defer rows.Close()

	var products []domain.Product

	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Stock,
			&product.Description,
			&product.Img,
			&product.CategoryID,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	// Always check rows.Err() after the loop
	return products, rows.Err()
}

// Save implements domain.ProductRepository.
func (p *productRepository) Save(ctx context.Context, product domain.Product) (domain.Product, error) {
	q := "INSERT INTO products (name, price, stock, description, img, category_id) VALUES (?, ?, ?, ?, ?, ?)"

	result, err := p.db.ExecContext(ctx, q, product.Name, product.Price, product.Stock, product.Description, product.Img, product.CategoryID)
	if err != nil {
		return domain.Product{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Product{}, err
	}

	product.ID = int(id)
	return product, nil
}

// Update implements domain.ProductRepository.
func (p *productRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	q := "UPDATE products SET name = ?, price = ?, stock = ?, description = ?, img = ?, category_id = ? WHERE id = ?"

	_, err := p.db.ExecContext(ctx, q, product.Name, product.Price, product.Stock, product.Description, product.Img, product.CategoryID, product.ID)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

// DecrementStock implements domain.ProductRepository.
func (p *productRepository) DecrementStock(ctx context.Context, tx *sql.Tx, productID int, qty int) error {
	q := `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`
	result, err := tx.ExecContext(ctx, q, qty, productID, qty)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrInvalidInput
	}
	return nil
}
