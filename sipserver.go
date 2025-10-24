package main

import (
	"bytes"
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
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	activeCallsMu sync.RWMutex
	activeCalls   = make(map[string]bool)
)

func StartSIPServer(ctx context.Context, cfg Config, debug bool) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received webhook: Method=%s, Headers=%v", r.Method, r.Header)
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

			activeCallsMu.Lock()
			if activeCalls[d.CallID] {
				activeCallsMu.Unlock()
				if debug {
					log.Printf("Call %s already accepted, ignoring duplicate webhook", d.CallID)
				}
				w.WriteHeader(200)
				return
			}
			activeCalls[d.CallID] = true
			activeCallsMu.Unlock()

			if err := acceptCall(cfg.APIKey, d.CallID, cfg.VoiceInstructions, cfg.Voice); err != nil {
				activeCallsMu.Lock()
				delete(activeCalls, d.CallID)
				activeCallsMu.Unlock()

				if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "call_id_not_found") {
					if debug {
						log.Printf("Call %s not found (possibly already ended or invalid): %v", d.CallID, err)
					}
					// Don't return error - just acknowledge the webhook
					w.WriteHeader(200)
					return
				}

				log.Printf("acceptCall: %v", err)
				http.Error(w, "accept failed", 500)
				return
			}
			go connectCallWS(cfg.APIKey, d.CallID, debug)
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
	s := strings.TrimPrefix(secret, "whsec_")
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
	req, err := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls/"+callID+"/accept",
		bytes.NewReader(b))
	if err != nil {
		return err
	}
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

func connectCallWS(apiKey, callID string, debug bool) {

	url := "wss://api.openai.com/v1/realtime?call_id=" + callID
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+apiKey)
	hdr.Set("Origin", "https://api.openai.com")
	d := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := d.Dial(url, hdr)
	if err != nil {
		log.Printf("ws dial: %v", err)
		return
	}
	defer func() {
		conn.Close()
		activeCallsMu.Lock()
		delete(activeCalls, callID)
		activeCallsMu.Unlock()
		if debug {
			log.Printf("Call %s cleanup complete", callID)
		}
	}()

	sessionUpdate := map[string]any{
		"type": "session.update",
		"session": map[string]any{
			"type":  "realtime",
			"tools": GetToolDefinitions(),
		},
	}
	if err := conn.WriteJSON(sessionUpdate); err != nil {
		log.Printf("session.update error: %v", err)
		return
	}

	if debug {
		log.Println("Session updated with tools")
	}

	greeting := map[string]any{
		"type": "response.create",
		"response": map[string]any{
			"instructions": "התחל את השיחה בברכה קצרה: 'שלום! הגעת לעוזר ניהול התורים. איך אפשר לעזור?'",
		},
	}
	_ = conn.WriteJSON(greeting)

	funcArgBuf := make(map[string]*strings.Builder)
	pendingFuncNames := make(map[string]string)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if debug {
			// log.Printf("WS: %s", string(msg))
		}
		if shouldEnd := handleVoiceEvent(conn, msg, funcArgBuf, pendingFuncNames, apiKey, callID, debug); shouldEnd {
			break
		}
	}
}

func handleVoiceEvent(conn *websocket.Conn, msg []byte, funcArgBuf map[string]*strings.Builder, pendingFuncNames map[string]string, apiKey, callID string, debug bool) bool {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &head); err != nil {
		log.Printf("json err: %v", err)
		return false
	}

	switch head.Type {
	case "response.output_item.added":
		var e map[string]any
		_ = json.Unmarshal(msg, &e)
		if item, ok := e["item"].(map[string]any); ok {
			if t, _ := item["type"].(string); t == "function_call" {
				callID, _ := item["call_id"].(string)
				name, _ := item["name"].(string)
				if callID != "" && name != "" {
					pendingFuncNames[callID] = name
					if debug {
						log.Printf("Function call detected: %s (callID: %s)", name, callID)
					}
				}
			}
		}

	case "response.function_call_arguments.delta":
		var e struct {
			Type   string `json:"type"`
			CallID string `json:"call_id"`
			Delta  string `json:"delta"`
		}
		_ = json.Unmarshal(msg, &e)
		if _, ok := funcArgBuf[e.CallID]; !ok {
			funcArgBuf[e.CallID] = &strings.Builder{}
		}
		funcArgBuf[e.CallID].WriteString(e.Delta)

	case "response.function_call_arguments.done":
		var e struct {
			Type   string `json:"type"`
			CallID string `json:"call_id"`
		}
		_ = json.Unmarshal(msg, &e)

		buf := ""
		if b, ok := funcArgBuf[e.CallID]; ok {
			buf = b.String()
		}
		fn := pendingFuncNames[e.CallID]

		if debug {
			log.Printf("Executing function: %s with args: %s", fn, buf)
		}

		if fn == "end_call" {
			log.Println("Assistant requested call end")

			sendFunctionResult(conn, e.CallID, map[string]any{
				"status":  "success",
				"message": "להתראות!",
			})

			_ = conn.WriteJSON(map[string]any{
				"type": "response.create",
				"response": map[string]any{
					"instructions": "אמור להתראות בצורה קצרה ונעימה",
				},
			})

			time.Sleep(2 * time.Second)

			if err := hangupCall(apiKey, callID); err != nil {
				log.Printf("Failed to hang up call: %v", err)
			} else {
				log.Println("Call hung up successfully via API")
			}

			conn.Close()
			return true
		}

		out := ExecuteVoiceFunction(fn, buf, debug)

		sendFunctionResult(conn, e.CallID, out)

		_ = conn.WriteJSON(map[string]any{
			"type": "response.create",
			"response": map[string]any{
				"instructions": "השתמש בתוצאת הפונקציה כדי לענות למשתמש בעברית טבעית.",
			},
		})

		delete(funcArgBuf, e.CallID)
		delete(pendingFuncNames, e.CallID)
	}
	return false
}

func sendFunctionResult(conn *websocket.Conn, callID string, output any) error {
	b, _ := json.Marshal(output)
	return conn.WriteJSON(map[string]any{
		"type": "conversation.item.create",
		"item": map[string]any{
			"type":    "function_call_output",
			"call_id": callID,
			"output":  string(b),
		},
	})
}
func hangupCall(apiKey, callID string) error {
	req, err := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls/"+callID+"/hangup",
		nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		out, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hangup %d: %s", resp.StatusCode, out)
	}
	return nil
}
