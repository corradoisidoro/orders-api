package order

import (
	"context"
	"errors"

	"github.com/corradoisidoro/orders-api/model"
)

// OrderRepository defines the contract for order persistence.
type OrderRepository interface {
	Insert(ctx context.Context, order *model.Order) error
	FindAll(ctx context.Context, page Page) (Result, error)
	FindByID(ctx context.Context, id int64) (model.Order, error)
	UpdateByID(ctx context.Context, order *model.Order) error
	DeleteByID(ctx context.Context, id int64) error
}

// Domain-level errors returned by the repository.
var (
	ErrNotExist     = errors.New("order does not exist")
	ErrInvalidInput = errors.New("invalid input provided")
)

// Page represents pagination parameters.
type Page struct {
	Size   int64 // number of items to return; 0 means "no limit"
	Offset int64 // cursor/offset for pagination; must be >= 0
}

// Result represents a paginated list of orders.
type Result struct {
	Orders []model.Order
	Cursor int64 // cursor for the next page
}

// Internal constants used across the repository.
const (
	orderIDColumn = "order_id"
)
