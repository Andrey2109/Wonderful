package main

import (
	"fmt"
	"log"
)

func main() {
	cfg := loadEnvVariables()
	fmt.Println("CLI starting… API key length:", len(cfg.APIKey))
	log.Println("OK: env loaded. Next: connect via WebSocket.")
}
