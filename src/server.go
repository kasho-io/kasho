package main

import (
	"log"
	"net/http"
	"sync"
)

type messageBroker struct {
	clients    map[chan []byte]struct{}
	mu         sync.RWMutex
	register   chan chan []byte
	unregister chan chan []byte
}

func newMessageBroker() *messageBroker {
	return &messageBroker{
		clients:    make(map[chan []byte]struct{}),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
	}
}

func (b *messageBroker) run() {
	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client] = struct{}{}
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			delete(b.clients, client)
			b.mu.Unlock()
			close(client)
		}
	}
}

func (b *messageBroker) broadcast(msg []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		select {
		case client <- msg:
		default:
			// Skip if client's channel is full
		}
	}
}

func startServer(broker *messageBroker) {
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		clientChan := make(chan []byte, 100)
		broker.register <- clientChan
		defer func() {
			broker.unregister <- clientChan
		}()

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case <-r.Context().Done():
				return
			case msg := <-clientChan:
				_, err := w.Write(append(msg, '\n'))
				if err != nil {
					return
				}
				flusher.Flush()
			}
		}
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
