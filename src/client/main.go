package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func prettyPrintJSON(data []byte) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(prettyJSON.String())
}

func main() {
	resp, err := http.Get("http://localhost:8080/stream")
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned status: %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	fmt.Println("Connected to server. Waiting for messages...")
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Fatalf("Error reading from stream: %v", err)
		}

		if len(line) > 0 {
			prettyPrintJSON(line)
		}
	}
}
