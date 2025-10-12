package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
			if err := acceptCall(cfg.APIKey, d.CallID, cfg.Instructions, cfg.Voice); err != nil {
				log.Printf("acceptCall: %v", err)
				http.Error(w, "accept failed", 500)
				return
			}
			w.Header().Set("Authorization", "Bearer "+cfg.APIKey)
			w.WriteHeader(200)
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

func acceptCall(apiKey, callID, instructions, voice string) error {
	body := map[string]any{
		"instructions": instructions,
		"type":         "realtime",
		"model":        "gpt-realtime",
		"audio": map[string]any{
			"output": map[string]any{"voice": voice},
		},
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls/"+callID+"/accept",
		io.NopCloser(bytesReader(b)))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		out, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("accept %d: %s", resp.StatusCode, out)
	}
	return nil
}

func bytesReader(b []byte) *os.File {
	f, _ := os.CreateTemp("", "b")
	f.Write(b)
	f.Seek(0, 0)
	return f
}
