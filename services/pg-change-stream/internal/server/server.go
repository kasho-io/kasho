package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"pg-change-stream/internal/types"
)

type messageBroker struct {
	clients    map[chan []byte]struct{}
	clientLSNs map[chan []byte]string
	mu         sync.RWMutex
	register   chan struct {
		ch  chan []byte
		lsn string
	}
	unregister chan chan []byte
	buffer     *RedisBuffer
}

func NewMessageBroker(redisURL string) (*messageBroker, error) {
	buffer, err := NewRedisBuffer(redisURL)
	if err != nil {
		return nil, err
	}

	return &messageBroker{
		clients:    make(map[chan []byte]struct{}),
		clientLSNs: make(map[chan []byte]string),
		register: make(chan struct {
			ch  chan []byte
			lsn string
		}),
		unregister: make(chan chan []byte),
		buffer:     buffer,
	}, nil
}

func (b *messageBroker) Run() {
	for {
		select {
		case reg := <-b.register:
			b.mu.Lock()
			b.clients[reg.ch] = struct{}{}
			b.clientLSNs[reg.ch] = reg.lsn
			b.mu.Unlock()
		case client := <-b.unregister:
			b.mu.Lock()
			delete(b.clients, client)
			delete(b.clientLSNs, client)
			b.mu.Unlock()
			close(client)
		}
	}
}

func (b *messageBroker) Broadcast(msg []byte, currentLSN string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		clientLSN := b.clientLSNs[client]
		if clientLSN == "" || clientLSN < currentLSN {
			select {
			case client <- msg:
			default:
				// Skip if client's channel is full
			}
		}
	}
}

func (mb *messageBroker) Close() error {
	return mb.buffer.Close()
}

func (mb *messageBroker) AddChange(ctx context.Context, lsn string, change types.Change) error {
	return mb.buffer.AddChange(ctx, lsn, change)
}

func StartServer(broker *messageBroker) {
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		lastLSN := r.URL.Query().Get("last_lsn")
		if lastLSN != "" {
			changes, err := broker.buffer.GetChangesAfter(r.Context(), lastLSN)
			if err != nil {
				log.Printf("Error getting buffered changes: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			for _, change := range changes {
				data, err := json.Marshal(change)
				if err != nil {
					log.Printf("Error marshaling change: %v", err)
					continue
				}
				w.Write(append(data, '\n'))
				w.(http.Flusher).Flush()
			}
		}

		clientChan := make(chan []byte, 100)
		broker.register <- struct {
			ch  chan []byte
			lsn string
		}{clientChan, lastLSN}
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
