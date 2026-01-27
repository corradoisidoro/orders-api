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

func NewOrderRepo(db *gorm.DB) OrderRepository {
	return &OrderRepo{DB: db}
}

// Insert creates a new order.
func (r *OrderRepo) Insert(ctx context.Context, order *model.Order) error {
	if err := validateOrderForInsert(order); err != nil {
		return err
	}

	if err := r.DB.WithContext(ctx).Create(order).Error; err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	return nil
}

// FindAll returns a page of orders and the cursor for the next page.
// If page.Size is 0, all remaining records are returned.
func (r *OrderRepo) FindAll(ctx context.Context, page Page) (Result, error) {
	if err := validatePage(page); err != nil {
		return Result{}, err
	}

	var orders []model.Order

	query := r.DB.WithContext(ctx).Offset(int(page.Offset))
	if page.Size > 0 {
		query = query.Limit(int(page.Size))
	}

	if err := query.Find(&orders).Error; err != nil {
		return Result{}, fmt.Errorf("find all orders: %w", err)
	}

	nextCursor := page.Offset + int64(len(orders))

	return Result{
		Orders: orders,
		Cursor: nextCursor,
	}, nil
}

// FindByID returns an order by its ID.
func (r *OrderRepo) FindByID(ctx context.Context, id int64) (model.Order, error) {
	if err := validateID(id); err != nil {
		return model.Order{}, err
	}

	var order model.Order
	result := r.DB.WithContext(ctx).
		Where(orderIDColumn+" = ?", id).
		First(&order)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.Order{}, fmt.Errorf("order %d: %w", id, ErrNotExist)
		}
		return model.Order{}, fmt.Errorf("find order %d: %w", id, result.Error)
	}

	return order, nil
}

// UpdateByID updates an existing order by its ID.
//
// NOTE: Updates(order) will overwrite zero-value fields.
// Callers should ensure the order struct contains the intended final state.
func (r *OrderRepo) UpdateByID(ctx context.Context, order *model.Order) error {
	if err := validateOrderForUpdate(order); err != nil {
		return err
	}

	result := r.DB.WithContext(ctx).
		Model(&model.Order{}).
		Where(orderIDColumn+" = ?", order.OrderID).
		Updates(order)

	if result.Error != nil {
		return fmt.Errorf("update order %d: %w", order.OrderID, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("update order %d: %w", order.OrderID, ErrNotExist)
	}

	return nil
}

// DeleteByID deletes an order by its ID.
func (r *OrderRepo) DeleteByID(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	result := r.DB.WithContext(ctx).
		Where(orderIDColumn+" = ?", id).
		Delete(&model.Order{})

	if result.Error != nil {
		return fmt.Errorf("delete order %d: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("delete order %d: %w", id, ErrNotExist)
	}

	return nil
}
