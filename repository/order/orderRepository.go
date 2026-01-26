package order

import (
	"context"

	"github.com/corradoisidoro/orders-api/model"
)

type OrderRepository interface {
	Insert(ctx context.Context, order *model.Order) error
	FindAll(ctx context.Context, page FindAllPage) (FindResult, error)
	FindByID(ctx context.Context, id int64) (model.Order, error)
	UpdateByID(ctx context.Context, order *model.Order) error
	DeleteByID(ctx context.Context, id int64) error
}
