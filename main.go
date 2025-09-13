package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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

	if err := InitializeSession(client); err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go client.readLoop(ctx)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type a message. Ctrl+C to exit.")
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down")
			return
		default:
		}

		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("stdin: %v", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if err := client.sendUserText(line); err != nil {
			log.Printf("send user text: %v", err)
			continue
		}
		if err := client.sendResponseCreate(""); err != nil {
			log.Printf("response.create: %v", err)
			continue
		}
	}
}
