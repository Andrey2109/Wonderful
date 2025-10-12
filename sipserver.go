package main

import (
	"context"
	"log"
	"net/http"
)

func StartSIPServer(ctx context.Context, cfg Config, debug bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(200)
	})

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	go func() {
		log.Printf("SIP webhook listening on http://0.0.0.0:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()
	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
}
