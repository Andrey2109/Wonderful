package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func setupOpenAIClient(cfg Config, model string, debug bool) (*WSClient, error) {
	url := "wss://api.openai.com/v1/realtime?model=" + model

	header := http.Header{}
	header.Set("Authorization", "Bearer "+cfg.APIKey)
	header.Set("OpenAI-Beta", "realtime=v1")

	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, resp, err := dialer.Dial(url, header)
	if err != nil {
		if resp != nil {
			log.Printf("handshake failed: status=%d", resp.StatusCode)
		}
		return nil, fmt.Errorf("dial failed: %v", err)
	}

	fmt.Println("Connected to", url)
	if debug {
		log.Println("WS connected, ready to send session")
	}

	client := &WSClient{
		Conn:             conn,
		Debug:            debug,
		Instructions:     cfg.Instructions,
		funcArgBuf:       map[string]*strings.Builder{},
		pendingFuncNames: map[string]string{},
	}

	return client, nil
}

func InitializeSession(client *WSClient) error {
	if err := client.sendSessionUpdate(); err != nil {
		return fmt.Errorf("session.update failed: %v", err)
	}
	log.Println("session.update sent")
	return nil
}
