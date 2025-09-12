package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Config struct {
	APIKey       string
	Instructions string
}
type WSClient struct {
	Conn         *websocket.Conn
	Debug        bool
	Instructions string
}

func loadEnvVariables() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := Config{
		APIKey:       os.Getenv("OPENAI_API_KEY"),
		Instructions: os.Getenv("INSTRUCTIONS"),
	}

	if config.APIKey == "" {
		log.Fatal("OPENAI_API_KEY is required")
	}
	if config.Instructions == "" {
		config.Instructions = "You are a concise CLI assistant."
	}
	return config
}
func (c *WSClient) writeJSON(v any) error {
	c.Conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	return c.Conn.WriteJSON(v)
}

func (c *WSClient) sendSessionUpdate() error {
	payload := map[string]any{
		"type": "session.update",
		"session": map[string]any{
			"modalities":   []string{"text"},
			"temperature":  0.7,
			"instructions": c.Instructions,
		},
	}
	return c.writeJSON(payload)
}
func (c *WSClient) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("read: %v", err)
			return
		}
		c.handleEvent(msg)
		if c.Debug {
			log.Printf("event: %s", string(msg))
		}
	}
}
func (c *WSClient) sendUserText(text string) error {
	item := map[string]any{
		"type": "message",
		"role": "user",
		"content": []any{
			map[string]any{"type": "input_text", "text": text},
		},
	}
	return c.writeJSON(map[string]any{
		"type": "conversation.item.create",
		"item": item,
	})
}

func (c *WSClient) sendResponseCreate(instructions string) error {
	resp := map[string]any{
		"modalities": []string{"text"},
	}
	if instructions != "" {
		resp["instructions"] = instructions
	}
	return c.writeJSON(map[string]any{
		"type":     "response.create",
		"response": resp,
	})
}
func (c *WSClient) handleEvent(msg []byte) {
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &head); err != nil {
		log.Printf("json err: %v", err)
		return
	}
	switch head.Type {
	case "response.text.delta":
		var e struct {
			Type  string `json:"type"`
			Delta string `json:"delta"`
		}
		_ = json.Unmarshal(msg, &e)
		fmt.Print(e.Delta)
	case "response.text.done":
		fmt.Println()
	default:
		if c.Debug {
			log.Printf("UNHANDLED: %s", string(msg))
		}
	}
}
