package httpx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter
	sendErr bool
}

func (w ResponseWriter) Error(code int, err error) error {
	http.Error(w, err.Error(), code)
	w.sendErr = false
	return err
}

func (w ResponseWriter) NoContent(code int) error {
	w.WriteHeader(code)
	return nil
}

func (w ResponseWriter) SqlcJson(code int, v any, err error) error {
	switch err {
	case nil:
		if v == nil {
			w.ResponseWriter.Header().Set("Content-Type", "application/json")
			w.WriteHeader(code)
			_, err := fmt.Fprint(w, "[]")
			return err
		}
		return w.Json(code, v)
	case sql.ErrNoRows:
		return w.NoContent(404)
	default:
		return w.Error(500, err)
	}
}

func (w ResponseWriter) Json(code int, v any) error {
	w.ResponseWriter.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w.ResponseWriter).Encode(v)
}

type HandlerFunc func(*ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wx := ResponseWriter{
		ResponseWriter: w,
		sendErr:        true,
	}
	err := f(&wx, r)
	if err != nil {
		if wx.sendErr {
			http.Error(wx, err.Error(), 500)
		}
		slog.Error(err.Error(), "url", r.URL.Path)
	}
}

type MiddlewareFunc func(next http.Handler) http.Handler

type ServeMux struct {
	*http.ServeMux
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		ServeMux: &http.ServeMux{},
	}
}

func (mux *ServeMux) HandleFunc(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	switch len(middlewares) {
	case 0:
		mux.Handle(pattern, handler)
	default:
		midhandler := middlewares[0](handler)
		for i, mid := range middlewares {
			if i == 0 {
				continue
			}
			midhandler = mid(midhandler)
		}
		mux.Handle(pattern, midhandler)
	}
}

func (mux *ServeMux) Use(middlewares ...MiddlewareFunc) http.Handler {
	midhandler := middlewares[0](mux.ServeMux)
	for i, mid := range middlewares {
		if i == 0 {
			continue
		}
		midhandler = mid(midhandler)
	}
	return midhandler
}
