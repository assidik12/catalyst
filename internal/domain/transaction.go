package domain

// Transaction represents the master record of a single order.
type Transaction struct {
	ID                  int
	TransactionDetailID string
	UserID              int
	TotalPrice          int
	Products            []TransactionDetail
}

// TransactionDetail represents one line-item inside a transaction.
type TransactionDetail struct {
	Quantity  int
	ProductID int
}
