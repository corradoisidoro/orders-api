package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/corradoisidoro/orders-api/internal/model"
	"github.com/corradoisidoro/orders-api/internal/repository"
	"github.com/stretchr/testify/assert"
)

//
// --- Repository error -> HTTP status mapping ---
//

// ErrInvalidInput is a caller error, so it must surface as 400 rather than 500.
// Previously every repository failure was reported as 500, which made validation
// problems look like server faults and tripped client-side alerting.
func TestOrderHandler_Create_InvalidInputIsBadRequest(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.InsertFn = func(ctx context.Context, o *model.Order) error {
		return fmt.Errorf("customer ID is required: %w", repository.ErrInvalidInput)
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodPost, "/orders", map[string]any{"customer_id": "1"})
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Update_InvalidInputIsBadRequest(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{OrderID: id, CustomerID: 1}, nil
	}
	mockRepo.UpdateByIDFn = func(ctx context.Context, o *model.Order) error {
		return fmt.Errorf("invalid order ID: %w", repository.ErrInvalidInput)
	}

	h := OrderHandler{Repo: mockRepo}

	req := withRouteParam(newRequest(http.MethodPatch, "/orders/1", map[string]string{"status": "shipped"}), "id", "1")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_List_InvalidInputIsBadRequest(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context, p repository.Page) (repository.Result, error) {
		return repository.Result{}, fmt.Errorf("invalid offset %d: %w", p.Offset, repository.ErrInvalidInput)
	}

	h := OrderHandler{Repo: mockRepo}

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_GetByID_NotFoundIs404(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{}, fmt.Errorf("order %d: %w", id, repository.ErrNotExist)
	}

	h := OrderHandler{Repo: mockRepo}

	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/orders/9", nil), "id", "9")
	rr := newRecorder()

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_DeleteByID_NotFoundIs404(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.DeleteByIDFn = func(ctx context.Context, id int64) error {
		return fmt.Errorf("order %d: %w", id, repository.ErrNotExist)
	}

	h := OrderHandler{Repo: mockRepo}

	req := withRouteParam(httptest.NewRequest(http.MethodDelete, "/orders/9", nil), "id", "9")
	rr := newRecorder()

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// An unrecognised repository failure is a genuine server fault and must stay a
// 500 with a generic client-facing message.
func TestOrderHandler_GetByID_UnexpectedErrorIs500(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{}, errors.New("connection reset by peer")
	}

	h := OrderHandler{Repo: mockRepo}

	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/orders/1", nil), "id", "1")
	rr := newRecorder()

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var resp map[string]string
	decodeResponseJSON(t, rr.Body.Bytes(), &resp)
	assert.Equal(t, "failed to retrieve order", resp["error"])
	assert.NotContains(t, resp["error"], "connection reset")
}

//
// --- Request body size limit ---
//

func TestOrderHandler_Create_BodyTooLargeIsRejected(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	// Comfortably larger than the 1 MiB cap.
	payload := `{"customer_id":"1","line_items":[{"quantity":1,"price":1,"pad":"` +
		strings.Repeat("x", 2<<20) + `"}]}`

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(payload))
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
