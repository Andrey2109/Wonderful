package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Config struct {
	APIKey       string
	Instructions string
}
type WSClient struct {
	Conn             *websocket.Conn
	Debug            bool
	Instructions     string
	funcArgBuf       map[string]*strings.Builder
	pendingFuncNames map[string]string
}

func readInstructionsFromFile(filename string) (string, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
func loadEnvVariables(debug bool) Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file %v", err)
	}

	config := Config{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}

	instructions, err := readInstructionsFromFile("instructions.txt")
	if err == nil && instructions != "" {
		config.Instructions = instructions
	}

	if debug {
		log.Printf("The instructions for the model are: %v", config.Instructions)
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
			"tools": []any{
				map[string]any{
					"type":        "function",
					"name":        "multiply",
					"description": "Multiply two numbers and return the product.",
					"parameters": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"a": map[string]any{"type": "number"},
							"b": map[string]any{"type": "number"},
						},
						"required": []string{"a", "b"},
					},
				},
			},
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

func (c *WSClient) sendFunctionResult(callID string, output any) error {
	b, _ := json.Marshal(output)
	return c.writeJSON(map[string]any{
		"type": "conversation.item.create",
		"item": map[string]any{
			"type":    "function_call_output",
			"call_id": callID,
			"output":  string(b),
		},
	})
}

func executeLocalFunction(name, argsJSON string) (any, error) {
	// fmt.Printf("\n=== FUNCTION CALL: %s ===\n", name)
	// fmt.Printf("Arguments: %s\n", argsJSON)
	if name != "multiply" {
		return map[string]any{"error": "unknown function", "name": name}, nil
	}
	var args struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return nil, fmt.Errorf("bad args: %w", err)
	}
	return map[string]any{"result": args.A * args.B}, nil
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
	case "response.output_item.added":
		var e map[string]any
		_ = json.Unmarshal(msg, &e)
		if item, ok := e["item"].(map[string]any); ok {
			if t, _ := item["type"].(string); t == "function_call" {
				callID, _ := item["call_id"].(string)
				name, _ := item["name"].(string)
				if callID != "" && name != "" {
					c.pendingFuncNames[callID] = name
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
		if _, ok := c.funcArgBuf[e.CallID]; !ok {
			c.funcArgBuf[e.CallID] = &strings.Builder{}
		}
		c.funcArgBuf[e.CallID].WriteString(e.Delta)
	case "response.function_call_arguments.done":
		var e struct {
			Type   string `json:"type"`
			CallID string `json:"call_id"`
		}
		_ = json.Unmarshal(msg, &e)
		buf := ""
		if b, ok := c.funcArgBuf[e.CallID]; ok {
			buf = b.String()
		}
		fn := c.pendingFuncNames[e.CallID]
		if fn == "" {
			fn = "multiply"
		}

		out, err := executeLocalFunction(fn, buf)
		if err != nil {
			out = map[string]any{"error": err.Error()}
		}

		_ = c.sendFunctionResult(e.CallID, out)
		_ = c.sendResponseCreate("Use the tool result to answer the user.")
	case "response.done":
		fmt.Println()

	default:
		if c.Debug {
			log.Printf("UNHANDLED: %s", string(msg))
		}
	}
}
