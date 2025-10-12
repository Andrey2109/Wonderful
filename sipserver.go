package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func StartSIPServer(ctx context.Context, cfg Config, debug bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		if !cfg.DisableWebhookSigCheck && cfg.WebhookSecret != "" {
			sig := r.Header.Get("webhook-signature")
			if sig == "" {
				sig = r.Header.Get("OpenAI-Signature")
			}
			if !verifyHMAC(raw, cfg.WebhookSecret, sig) {
				http.Error(w, "Invalid signature", http.StatusBadRequest)
				return
			}
		}
		var ev struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(raw, &ev); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if ev.Type == "realtime.call.incoming" {
			var d struct {
				CallID string `json:"call_id"`
			}
			_ = json.Unmarshal(ev.Data, &d)
			if d.CallID == "" {
				http.Error(w, "no call_id", 400)
				return
			}
			// TODO: accept call
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

func verifyHMAC(body []byte, secret, provided string) bool {
	s := secret
	if strings.HasPrefix(s, "whsec_") {
		s = s[len("whsec_"):]
	}
	h := hmac.New(sha256.New, []byte(s))
	h.Write(body)
	want := hex.EncodeToString(h.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(want), []byte(strings.TrimSpace(provided))) == 1
}
