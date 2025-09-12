package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const defaultModel = "gpt-4o-mini-realtime-preview-2024-12-17"

func main() {
	cfg := loadEnvVariables()
	model := flag.String("model", defaultModel, "OpenAI Realtime model id")
	debug := flag.Bool("debug", false, "print raw events")
	flag.Parse()

	url := "wss://api.openai.com/v1/realtime?model=" + *model

	header := http.Header{}
	header.Set("Authorization", "Bearer "+cfg.APIKey)
	header.Set("OpenAI-Beta", "realtime=v1")

	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, resp, err := dialer.Dial(url, header)
	if err != nil {
		if resp != nil {
			log.Printf("handshake failed: status=%d", resp.StatusCode)
		}
		log.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Println("Connected to", url)
	if *debug {
		log.Println("WS connected, ready to send session")
	}

	client := &WSClient{Conn: conn, Debug: *debug, Instructions: cfg.Instructions}

	if err := client.sendSessionUpdate(); err != nil {
		log.Fatalf("session.update failed: %v", err)
	}
	log.Println("session.update sent")
}
