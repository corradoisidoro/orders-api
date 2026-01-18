package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/corradoisidoro/orders-api/model"
	"gorm.io/gorm"
)

type OrderRepo struct {
	DB *gorm.DB
}

var ErrNotExist = errors.New("order does not exist")

type FindAllPage struct {
	Size   uint64 // page size
	Offset uint64 // cursor/offset
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func NewOrderRepo(db *gorm.DB) OrderRepository {
	return &OrderRepo{DB: db}
}

func (r *OrderRepo) Insert(ctx context.Context, order *model.Order) error {
	if err := r.DB.WithContext(ctx).Create(&order).Error; err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

func (r *OrderRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	var orders []model.Order

	query := r.DB.WithContext(ctx).Offset(int(page.Offset))
	if page.Size > 0 {
		query = query.Limit(int(page.Size))
	}

	if err := query.Find(&orders).Error; err != nil {
		return FindResult{}, fmt.Errorf("retrieve orders: %w", err)
	}

	nextCursor := page.Offset + uint64(len(orders))

	return FindResult{
		Orders: orders,
		Cursor: nextCursor,
	}, nil
}

func (r *OrderRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	var order model.Order

	result := r.DB.WithContext(ctx).First(&order, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.Order{}, ErrNotExist
		}
		return model.Order{}, fmt.Errorf("retrieve order %d: %w", id, result.Error)
	}

	return order, nil
}

func (r *OrderRepo) UpdateByID(ctx context.Context, order *model.Order) error {
	result := r.DB.WithContext(ctx).Save(order)
	if result.Error != nil {
		return fmt.Errorf("update order %d: %w", order.OrderID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update order %d: %w", order.OrderID, ErrNotExist)
	}
	return nil
}

func (r *OrderRepo) DeleteByID(ctx context.Context, id uint64) error {
	result := r.DB.WithContext(ctx).Delete(&model.Order{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete order %d: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrNotExist
	}
	return nil
}
