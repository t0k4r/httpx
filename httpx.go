package httpx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func Error(w http.ResponseWriter, err error, code int) error {
	http.Error(w, err.Error(), code)
	return nil
}

func Empty(w http.ResponseWriter, code int) error {
	w.WriteHeader(code)
	return nil
}
func Json(w http.ResponseWriter, v any, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}
func JsonMany[T any](w http.ResponseWriter, v []T, code int) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)
	switch len(v) {
	case 0:
		_, err := fmt.Fprint(w, "[]")
		return err
	default:
		return json.NewEncoder(w).Encode(v)
	}
}
func JsonSql(w http.ResponseWriter, v any, err error, code int) error {
	switch err {
	case nil:
		return Json(w, v, code)
	case sql.ErrNoRows:
		return Empty(w, 404)
	default:
		return err
	}
}
func JsonSqlMany[T any](w http.ResponseWriter, v []T, err error, code int) error {
	switch err {
	case nil:
		return JsonMany(w, v, code)
	case sql.ErrNoRows:
		return Empty(w, 404)
	default:
		return err
	}
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type MiddlewareFunc func(http.Handler) http.Handler

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		slog.Error(err.Error(), "type", "error", "parh", r.URL.Path)
		http.Error(w, err.Error(), 500)
	}
}

type ServeMux struct {
	*http.ServeMux
	middlewares []MiddlewareFunc
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		ServeMux:    &http.ServeMux{},
		middlewares: []MiddlewareFunc{},
	}
}

func (mux *ServeMux) Use(middlewares ...MiddlewareFunc) {
	mux.middlewares = append(mux.middlewares, middlewares...)
}
func (mux *ServeMux) wrap(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for _, mid := range append(mux.middlewares, middlewares...) {
		handler = mid(handler)
	}
	return handler
}

func (mux *ServeMux) Handle(pattern string, handler http.Handler, middlewares ...MiddlewareFunc) {
	mux.ServeMux.Handle(pattern, mux.wrap(handler))
}
func (mux *ServeMux) HandleFunc(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	mux.ServeMux.Handle(pattern, mux.wrap(handler))
}

func (mux *ServeMux) Handlex(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	mux.ServeMux.Handle(pattern, mux.wrap(handler))
}
