package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"requestId"`
}
type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error *Error `json:"error,omitempty"`
	Meta  any    `json:"meta,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: data})
}

func Fail(w http.ResponseWriter, r *http.Request, status int, code, message string, details any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Error: &Error{Code: code, Message: message, Details: details, RequestID: RequestID(r.Context())}})
}

func Decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		Fail(w, r, http.StatusBadRequest, "invalid_request", "请求格式无效", nil)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		Fail(w, r, http.StatusBadRequest, "invalid_request", "请求格式无效", nil)
		return false
	}
	return true
}

var ErrNotFound = errors.New("not found")
