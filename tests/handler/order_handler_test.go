package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/corradoisidoro/orders-api/handler"
	"github.com/corradoisidoro/orders-api/model"
	"github.com/corradoisidoro/orders-api/repository/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// --- Mock Repository ---
//

type mockOrderRepo struct {
	InsertFn     func(ctx context.Context, o *model.Order) error
	FindAllFn    func(ctx context.Context, p order.Page) (order.Result, error)
	FindByIDFn   func(ctx context.Context, id int64) (model.Order, error)
	UpdateByIDFn func(ctx context.Context, o *model.Order) error
	DeleteByIDFn func(ctx context.Context, id int64) error
}

func (m *mockOrderRepo) Insert(ctx context.Context, o *model.Order) error {
	return m.InsertFn(ctx, o)
}
func (m *mockOrderRepo) FindAll(ctx context.Context, p order.Page) (order.Result, error) {
	return m.FindAllFn(ctx, p)
}
func (m *mockOrderRepo) FindByID(ctx context.Context, id int64) (model.Order, error) {
	return m.FindByIDFn(ctx, id)
}
func (m *mockOrderRepo) UpdateByID(ctx context.Context, o *model.Order) error {
	return m.UpdateByIDFn(ctx, o)
}
func (m *mockOrderRepo) DeleteByID(ctx context.Context, id int64) error {
	return m.DeleteByIDFn(ctx, id)
}

//
// --- Helpers ---
//

func newRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	return req
}

func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

//
// --- CREATE ---
//

func TestOrderHandler_Create_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		InsertFn: func(ctx context.Context, o *model.Order) error {
			o.OrderID = 123
			return nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]any{
		"customer_id": "1",
		"line_items":  []model.LineItem{},
	}

	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp model.Order
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, int64(123), resp.OrderID)
}

func TestOrderHandler_Create_InvalidJSON(t *testing.T) {
	mockRepo := &mockOrderRepo{}
	h := handler.Order{Repo: mockRepo}

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("{invalid"))
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_InvalidCustomerID(t *testing.T) {
	mockRepo := &mockOrderRepo{}
	h := handler.Order{Repo: mockRepo}

	body := map[string]any{"customer_id": 0}
	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_InsertFails(t *testing.T) {
	mockRepo := &mockOrderRepo{
		InsertFn: func(ctx context.Context, o *model.Order) error {
			return errors.New("db error")
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]any{"customer_id": "1"}
	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- LIST ---
//

func TestOrderHandler_List_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindAllFn: func(ctx context.Context, p order.Page) (order.Result, error) {
			return order.Result{
				Orders: []model.Order{
					{OrderID: 1, LineItems: nil},
				},
				Cursor: 10,
			}, nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders?cursor=0", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Items []model.Order `json:"items"`
		Next  int64         `json:"next"`
	}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

	assert.Equal(t, int64(10), resp.Next)
	assert.Equal(t, []model.LineItem{}, resp.Items[0].LineItems)
}

func TestOrderHandler_List_InvalidCursor(t *testing.T) {
	h := handler.Order{Repo: &mockOrderRepo{}}

	req := newRequest(http.MethodGet, "/orders?cursor=abc", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_List_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindAllFn: func(ctx context.Context, p order.Page) (order.Result, error) {
			return order.Result{}, errors.New("db error")
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders?cursor=0", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- GET BY ID ---
//

func TestOrderHandler_GetByID_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{OrderID: id, LineItems: nil}, nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders/5", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_GetByID_InvalidID(t *testing.T) {
	h := handler.Order{Repo: &mockOrderRepo{}}

	req := newRequest(http.MethodGet, "/orders/abc", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{}, order.ErrNotExist
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders/5", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_GetByID_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{}, errors.New("db error")
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders/5", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- UPDATE ---
//

func TestOrderHandler_UpdateByID_Shipped_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{OrderID: id}, nil
		},
		UpdateByIDFn: func(ctx context.Context, o *model.Order) error {
			return nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]string{"status": "shipped"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_UpdateByID_Completed_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			now := time.Now().UTC()
			return model.Order{OrderID: id, ShippedAt: &now}, nil
		},
		UpdateByIDFn: func(ctx context.Context, o *model.Order) error {
			return nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]string{"status": "completed"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_UpdateByID_InvalidJSON(t *testing.T) {
	h := handler.Order{Repo: &mockOrderRepo{}}

	req := httptest.NewRequest(http.MethodPatch, "/orders/5", bytes.NewBufferString("{invalid"))
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateByID_InvalidStatus(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{OrderID: id}, nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]string{"status": "unknown"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateByID_NotFound(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{}, order.ErrNotExist
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]string{"status": "shipped"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_UpdateByID_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		FindByIDFn: func(ctx context.Context, id int64) (model.Order, error) {
			return model.Order{OrderID: id}, nil
		},
		UpdateByIDFn: func(ctx context.Context, o *model.Order) error {
			return errors.New("db error")
		},
	}

	h := handler.Order{Repo: mockRepo}

	body := map[string]string{"status": "shipped"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- DELETE ---
//

func TestOrderHandler_DeleteByID_Success(t *testing.T) {
	mockRepo := &mockOrderRepo{
		DeleteByIDFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodDelete, "/orders/5", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestOrderHandler_DeleteByID_NotFound(t *testing.T) {
	mockRepo := &mockOrderRepo{
		DeleteByIDFn: func(ctx context.Context, id int64) error {
			return order.ErrNotExist
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodDelete, "/orders/5", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_DeleteByID_RepoError(t *testing.T) {
	mockRepo := &mockOrderRepo{
		DeleteByIDFn: func(ctx context.Context, id int64) error {
			return errors.New("db error")
		},
	}

	h := handler.Order{Repo: mockRepo}

	req := newRequest(http.MethodDelete, "/orders/5", nil)
	rr := newRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
