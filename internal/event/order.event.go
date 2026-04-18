package event

import (
	"context"
	"time"
)

const (
	TopicOrderCreated = "order.created"
)

// Producer is the interface for event publishing.
type Producer interface {
	Publish(ctx context.Context, topic string, data any) error
}

type ProductItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// OrderCreatedEvent represents the data published to Kafka when a transaction is completed.
type OrderCreatedEvent struct {
	TransactionID string        `json:"transaction_id"`
	UserID        int           `json:"user_id"`
	TotalPrice    int           `json:"total_price"`
	Products      []ProductItem `json:"products"`
	CreatedAt     time.Time     `json:"created_at"`
}
