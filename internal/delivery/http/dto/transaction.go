package dto

type TransactionItem struct {
	ID  int `json:"id"`
	Qty int `json:"qty"`
}

type TransactionRequest struct {
	Products []TransactionItem `json:"products"`
}

type TransactionResponse struct {
	ID         string `json:"id"`
	TotalPrice int    `json:"totalPrice"`
	Products   []struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
		Qty   int    `json:"qty"`
	} `json:"products"`
}
