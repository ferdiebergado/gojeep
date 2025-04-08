package handler

import (
	"log"
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
		log.Println("pprof listening on :" + port)
		if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()
}
