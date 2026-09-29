package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// maxBodyBytes caps the size of a decoded request body. Without it, a client can
// make the process buffer an arbitrarily large payload.
const maxBodyBytes = 1 << 20 // 1 MiB

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid id", err)
		return 0, false
	}
	return id, true
}

func parseQueryInt(w http.ResponseWriter, r *http.Request, key string, defaultVal int64) (int64, bool) {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return defaultVal, true
	}

	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid "+key, err)
		return 0, false
	}

	return val, true
}

// decodeJSON reads a size-capped JSON body. A body that exceeds maxBodyBytes
// (or is otherwise malformed) is reported as a 400.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError sends a generic error to the client. For 5xx responses the cause is
// logged rather than returned, so a client-facing message never leaks internals
// while the detail is still recoverable from the logs.
func writeError(w http.ResponseWriter, r *http.Request, status int, msg string, cause error) {
	if status >= http.StatusInternalServerError {
		log.Printf("error: %s %s: %s: %v", r.Method, r.URL.Path, msg, cause)
	} else if cause != nil {
		log.Printf("warn: %s %s: %s: %v", r.Method, r.URL.Path, msg, cause)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}
