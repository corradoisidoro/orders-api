package order

import (
	"fmt"

	"github.com/corradoisidoro/orders-api/model"
)

// validateID ensures the ID is positive.
func validateID(id int64) error {
	if id <= 0 {
		return fmt.Errorf("invalid ID %d: %w", id, ErrInvalidInput)
	}
	return nil
}

// validatePage ensures pagination parameters are valid.
func validatePage(page Page) error {
	if page.Offset < 0 {
		return fmt.Errorf("invalid offset %d: %w", page.Offset, ErrInvalidInput)
	}
	if page.Size < 0 {
		return fmt.Errorf("invalid size %d: %w", page.Size, ErrInvalidInput)
	}
	return nil
}

// validateOrderForInsert ensures the order is valid for creation.
func validateOrderForInsert(order *model.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil: %w", ErrInvalidInput)
	}
	if order.CustomerID < 1 {
		return fmt.Errorf("customer ID is required for insert: %w", ErrInvalidInput)
	}
	return nil
}

// validateOrderForUpdate ensures the order is valid for updating.
func validateOrderForUpdate(order *model.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil: %w", ErrInvalidInput)
	}
	if order.OrderID <= 0 {
		return fmt.Errorf("invalid order ID %d for update: %w", order.OrderID, ErrInvalidInput)
	}
	if order.CustomerID < 1 {
		return fmt.Errorf("customer ID is required for update: %w", ErrInvalidInput)
	}
	return nil
}
