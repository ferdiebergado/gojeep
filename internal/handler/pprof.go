package handler

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func StartPProf() {
	if os.Getenv("PPROF") != "true" {
		return
	}

	go func() {
		const port = "6060"
		slog.Info("pprof listening", "port", port)
		if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
			slog.Error("pprof server error", "reason", err)
		}
	}()
}
