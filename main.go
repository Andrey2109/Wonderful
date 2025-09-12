package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	client := &WSClient{Conn: conn,
		Debug:            *debug,
		Instructions:     cfg.Instructions,
		funcArgBuf:       map[string]*strings.Builder{},
		pendingFuncNames: map[string]string{}}

	if err := client.sendSessionUpdate(); err != nil {
		log.Fatalf("session.update failed: %v", err)
	}
	log.Println("session.update sent")

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
