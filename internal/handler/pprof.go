package handler

import (
	"net/http/pprof"

	"github.com/ferdiebergado/goexpress"
)

func MountPProfRoutes(router *goexpress.Router) {
	router.Get("/debug/pprof/", pprof.Index)
	router.Get("/debug/pprof/cmdline", pprof.Cmdline)
	router.Get("/debug/pprof/profile", pprof.Profile)
	router.Get("/debug/pprof/symbol", pprof.Symbol)
	router.Get("/debug/pprof/trace", pprof.Trace)
	router.Handle("GET /debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("GET /debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("GET /debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("GET /debug/pprof/block", pprof.Handler("block"))
}
