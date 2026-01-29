package handler

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

	"github.com/corradoisidoro/orders-api/internal/model"
	"github.com/corradoisidoro/orders-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// --- Mock Repository ---
//

type mockOrderRepo struct {
	InsertFn     func(ctx context.Context, o *model.Order) error
	FindAllFn    func(ctx context.Context, p repository.Page) (repository.Result, error)
	FindByIDFn   func(ctx context.Context, id int64) (model.Order, error)
	UpdateByIDFn func(ctx context.Context, o *model.Order) error
	DeleteByIDFn func(ctx context.Context, id int64) error
}

func newMockRepo() *mockOrderRepo {
	return &mockOrderRepo{
		InsertFn: func(ctx context.Context, o *model.Order) error { return nil },
		FindAllFn: func(ctx context.Context, p repository.Page) (repository.Result, error) {
			return repository.Result{}, nil
		},
		FindByIDFn:   func(ctx context.Context, id int64) (model.Order, error) { return model.Order{}, nil },
		UpdateByIDFn: func(ctx context.Context, o *model.Order) error { return nil },
		DeleteByIDFn: func(ctx context.Context, id int64) error { return nil },
	}
}

func (m *mockOrderRepo) Insert(ctx context.Context, o *model.Order) error {
	return m.InsertFn(ctx, o)
}
func (m *mockOrderRepo) FindAll(ctx context.Context, p repository.Page) (repository.Result, error) {
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
	return httptest.NewRequest(method, path, &buf)
}

func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

func withRouteParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func decodeResponseJSON(t *testing.T, body []byte, v any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(body, v))
}

//
// --- CREATE ---
//

func TestOrderHandler_Create_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.InsertFn = func(ctx context.Context, o *model.Order) error {
		o.OrderID = 123
		return nil
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]any{
		"customer_id": "1",
		"line_items":  []model.LineItem{},
	}

	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp model.Order
	decodeResponseJSON(t, rr.Body.Bytes(), &resp)
	assert.Equal(t, int64(123), resp.OrderID)
}

func TestOrderHandler_Create_InvalidJSON(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString("{invalid"))
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_InvalidCustomerID(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	body := map[string]any{"customer_id": 0}
	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_Create_InsertFails(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.InsertFn = func(ctx context.Context, o *model.Order) error {
		return errors.New("db error")
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]any{"customer_id": "1"}
	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestOrderHandler_Create_MissingCustomerID(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	body := map[string]any{}
	req := newRequest(http.MethodPost, "/orders", body)
	rr := newRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

//
// --- LIST ---
//

func TestOrderHandler_List_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context, p repository.Page) (repository.Result, error) {
		return repository.Result{
			Orders: []model.Order{
				{OrderID: 1, LineItems: nil},
			},
			Cursor: 10,
		}, nil
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders?cursor=0", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Items []model.Order `json:"items"`
		Next  int64         `json:"next"`
	}
	decodeResponseJSON(t, rr.Body.Bytes(), &resp)

	assert.Equal(t, int64(10), resp.Next)
	assert.Equal(t, []model.LineItem{}, resp.Items[0].LineItems)
}

func TestOrderHandler_List_InvalidCursor(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	req := newRequest(http.MethodGet, "/orders?cursor=abc", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_List_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindAllFn = func(ctx context.Context, p repository.Page) (repository.Result, error) {
		return repository.Result{}, errors.New("db error")
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders?cursor=0", nil)
	rr := newRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- GET BY ID ---
//

func TestOrderHandler_GetByID_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{OrderID: id, LineItems: nil}, nil
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_GetByID_InvalidID(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	req := newRequest(http.MethodGet, "/orders/abc", nil)
	req = withRouteParam(req, "id", "abc")
	rr := newRecorder()

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{}, repository.ErrNotExist
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_GetByID_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{}, errors.New("db error")
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodGet, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- UPDATE ---
//

func TestOrderHandler_UpdateByID_Shipped_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{OrderID: id}, nil
	}
	mockRepo.UpdateByIDFn = func(ctx context.Context, o *model.Order) error {
		return nil
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]string{"status": "shipped"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_UpdateByID_Completed_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		now := time.Now().UTC()
		return model.Order{OrderID: id, ShippedAt: &now}, nil
	}
	mockRepo.UpdateByIDFn = func(ctx context.Context, o *model.Order) error {
		return nil
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]string{"status": "completed"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrderHandler_UpdateByID_InvalidJSON(t *testing.T) {
	h := OrderHandler{Repo: newMockRepo()}

	req := httptest.NewRequest(http.MethodPatch, "/orders/5", bytes.NewBufferString("{invalid"))
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateByID_InvalidStatus(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{OrderID: id}, nil
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]string{"status": "unknown"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrderHandler_UpdateByID_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{}, repository.ErrNotExist
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]string{"status": "shipped"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_UpdateByID_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.FindByIDFn = func(ctx context.Context, id int64) (model.Order, error) {
		return model.Order{OrderID: id}, nil
	}
	mockRepo.UpdateByIDFn = func(ctx context.Context, o *model.Order) error {
		return errors.New("db error")
	}

	h := OrderHandler{Repo: mockRepo}

	body := map[string]string{"status": "shipped"}
	req := newRequest(http.MethodPatch, "/orders/5", body)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.UpdateByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

//
// --- DELETE ---
//

func TestOrderHandler_DeleteByID_Success(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.DeleteByIDFn = func(ctx context.Context, id int64) error {
		return nil
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodDelete, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestOrderHandler_DeleteByID_NotFound(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.DeleteByIDFn = func(ctx context.Context, id int64) error {
		return repository.ErrNotExist
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodDelete, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestOrderHandler_DeleteByID_RepoError(t *testing.T) {
	mockRepo := newMockRepo()
	mockRepo.DeleteByIDFn = func(ctx context.Context, id int64) error {
		return errors.New("db error")
	}

	h := OrderHandler{Repo: mockRepo}

	req := newRequest(http.MethodDelete, "/orders/5", nil)
	req = withRouteParam(req, "id", "5")
	rr := newRecorder()

	h.DeleteByID(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
