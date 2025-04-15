package handler

import (
	"net/http"

	"github.com/ferdiebergado/goexpress"
)

type Router interface {
	http.Handler
	Use(middleware func(next http.Handler) http.Handler)
	Get(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler)
	Post(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler)
	Put(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler)
	Patch(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler)
	Delete(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler)
	Options(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler)
	Group(pattern string, groupFunc func(Router) Router, middlewares ...func(next http.Handler) http.Handler)
}

type router struct {
	handler *goexpress.Router
}

var _ Router = (*router)(nil)

func NewRouter() Router {
	return &router{
		handler: goexpress.New(),
	}
}

func (r *router) Delete(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handler.Delete(pattern, handlerFunc, middlewares...)
}

func (r *router) Get(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handler.Get(pattern, handlerFunc, middlewares...)
}

func (r *router) Group(prefix string, grpFunc func(Router) Router, middlewares ...func(next http.Handler) http.Handler) {
	grpHandler := grpFunc(NewRouter())

	r.handler.Handle(prefix+"/", http.StripPrefix(prefix, grpHandler), middlewares...)
}

func (r *router) Options(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handler.Options(pattern, handlerFunc, middlewares...)
}

func (r *router) Patch(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handler.Patch(pattern, handlerFunc, middlewares...)
}

func (r *router) Post(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handler.Post(pattern, handlerFunc, middlewares...)
}

func (r *router) Put(pattern string, handlerFunc http.HandlerFunc, middlewares ...func(next http.Handler) http.Handler) {
	r.handler.Put(pattern, handlerFunc, middlewares...)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *router) Use(middleware func(next http.Handler) http.Handler) {
	r.handler.Use(middleware)
}
