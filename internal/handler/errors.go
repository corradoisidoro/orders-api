package handler

import (
	"errors"
	"net/http"

	"github.com/corradoisidoro/orders-api/internal/repository"
)

// statusFor maps a repository error onto an HTTP status code. It returns 0 when
// the error has no specific mapping and the caller should fall back to 500.
func statusFor(err error) int {
	switch {
	case errors.Is(err, repository.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, repository.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return 0
	}
}

// writeRepoError responds to a repository failure, distinguishing caller errors
// (bad input, missing record) from genuine server faults.
func writeRepoError(w http.ResponseWriter, r *http.Request, err error, msg string) {
	if status := statusFor(err); status != 0 {
		writeError(w, r, status, msg, err)
		return
	}
	writeError(w, r, http.StatusInternalServerError, msg, err)
}
