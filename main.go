package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const defaultModel = "gpt-realtime"

func main() {
	mode := flag.String("mode", "cli", "cli | sip")
	model := flag.String("model", defaultModel, "OpenAI Realtime model id")
	debug := flag.Bool("debug", false, "print raw events")
	flag.Parse()
	cfg := loadEnvVariables(*debug, *mode)

	client, err := setupOpenAIClient(cfg, *model, *debug)
	if err != nil {
		log.Fatalf("failed to setup OpenAI client: %v", err)
	}
	defer client.Conn.Close()

	if err := InitializeSession(client, *debug); err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *mode == "sip" {
		if err := InitDB(cfg); err != nil {
			log.Fatalf("failed to initialize database: %v", err)
		}
		defer CloseDB()

		StartSIPServer(ctx, cfg, client.Debug)
		return
	}
	RunClientLoop(ctx, client)
}
