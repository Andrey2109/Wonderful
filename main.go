package main

import (
	"flag"
	"fmt"
	"log"
)

const defaultModel = "gpt-4o-mini-realtime-preview-pt-realtime-2025-08-28"

func main() {
	cfg := loadEnvVariables()
	_ = cfg
	model := flag.String("model", defaultModel, "OpenAI Realtime model id")
	debug := flag.Bool("debug", false, "print raw events")
	flag.Parse()

	fmt.Println("Model:", *model)
	if *debug {
		log.Println("Debug mode enabled")
	}
	log.Println("Env OK")
}
