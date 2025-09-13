package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const defaultModel = "gpt-4o-mini-realtime-preview-2024-12-17"

func main() {
	cfg := loadEnvVariables()
	model := flag.String("model", defaultModel, "OpenAI Realtime model id")
	debug := flag.Bool("debug", false, "print raw events")
	flag.Parse()

	client, err := setupOpenAIClient(cfg, *model, *debug)
	if err != nil {
		log.Fatalf("failed to setup OpenAI client: %v", err)
	}
	defer client.Conn.Close()

	if err := InitializeSession(client); err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	RunClientLoop(ctx, client)
}
