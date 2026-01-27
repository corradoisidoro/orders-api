package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/corradoisidoro/orders-api/model"
	"github.com/corradoisidoro/orders-api/repository/order"
)

type Order struct {
	Repo order.OrderRepository
}

func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID int64            `json:"customer_id,string"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if body.CustomerID <= 0 {
		writeError(w, http.StatusBadRequest, "customer_id must be > 0")
		return
	}

	now := time.Now().UTC()

	o := model.Order{
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	if err := h.Repo.Insert(r.Context(), &o); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create order")
		return
	}

	// Normalize line_items
	if o.LineItems == nil {
		o.LineItems = []model.LineItem{}
	}

	writeJSON(w, http.StatusCreated, o)
}

func (h *Order) List(w http.ResponseWriter, r *http.Request) {
	cursor, ok := parseQueryInt(w, r, "cursor", 0)
	if !ok {
		return
	}

	const defaultPageSize = 50

	res, err := h.Repo.FindAll(r.Context(), order.Page{
		Offset: cursor,
		Size:   defaultPageSize,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list orders")
		return
	}

	for i := range res.Orders {
		if res.Orders[i].LineItems == nil {
			res.Orders[i].LineItems = []model.LineItem{}
		}
	}

	response := struct {
		Items []model.Order `json:"items"`
		Next  int64         `json:"next"`
	}{
		Items: res.Orders,
		Next:  res.Cursor,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	o, err := h.Repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrNotExist) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve order")
		return
	}

	if o.LineItems == nil {
		o.LineItems = []model.LineItem{}
	}

	writeJSON(w, http.StatusOK, o)
}

func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	id, ok := parseID(w, r)
	if !ok {
		return
	}

	o, err := h.Repo.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrNotExist) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to retrieve order")
		return
	}

	now := time.Now().UTC()

	switch body.Status {
	case "shipped":
		if o.ShippedAt != nil {
			writeError(w, http.StatusBadRequest, "order already shipped")
			return
		}
		o.ShippedAt = &now

	case "completed":
		if o.ShippedAt == nil {
			writeError(w, http.StatusBadRequest, "order must be shipped before completion")
			return
		}
		if o.CompletedAt != nil {
			writeError(w, http.StatusBadRequest, "order already completed")
			return
		}
		o.CompletedAt = &now

	default:
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}

	if err := h.Repo.UpdateByID(r.Context(), &o); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update order")
		return
	}

	writeJSON(w, http.StatusOK, o)
}

func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	err := h.Repo.DeleteByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, order.ErrNotExist) {
			writeError(w, http.StatusNotFound, "order not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete order")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
