package httpx

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

var ErrHandler ErrHandlerFunc = defaultErrHandler

func defaultErrHandler(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		slog.Error(err.Error(), "path", r.URL.Path)
		http.Error(w, err.Error(), 500)
	}
}

func HtmlTempl(w http.ResponseWriter, templ *template.Template, data any, code int) error {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/html")
	err := templ.Execute(w, data)
	return err
}

func Html(w http.ResponseWriter, html string, code int) error {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/html")
	_, err := fmt.Fprint(w, html)
	return err

}

func Error(w http.ResponseWriter, err error, code int) error {
	http.Error(w, err.Error(), code)
	return nil
}

func NoContent(w http.ResponseWriter, code int) error {
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

type ErrHandlerFunc func(http.ResponseWriter, *http.Request, error)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type MiddlewareFunc func(http.Handler) http.Handler

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ErrHandler(w, r, f(w, r))
}

func wrap(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for _, mid := range middlewares {
		handler = mid(handler)
	}
	return handler
}

type ServeMux struct {
	*http.ServeMux
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		ServeMux: &http.ServeMux{},
	}
}

func (mux *ServeMux) With(middlewares ...MiddlewareFunc) http.Handler {
	return wrap(mux.ServeMux, middlewares...)
}

func (mux *ServeMux) Handle(pattern string, handler http.Handler, middlewares ...MiddlewareFunc) {
	mux.ServeMux.Handle(pattern, wrap(handler, middlewares...))
}
func (mux *ServeMux) HandleFunc(pattern string, handler http.HandlerFunc, middlewares ...MiddlewareFunc) {
	mux.ServeMux.Handle(pattern, wrap(handler, middlewares...))
}

func (mux *ServeMux) Handlex(pattern string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	mux.ServeMux.Handle(pattern, wrap(handler, middlewares...))
}
