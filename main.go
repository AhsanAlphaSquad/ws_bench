package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {

	connections := atomic.Int64{}

	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			log.Printf("connections: %d", connections.Load())
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("upgrade error: %v", err)
		}

		connections.Add(1)

		go func() {
			defer connections.Add(-1)
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					log.Printf("read error: %v", err)
					break
				}
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					log.Printf("write error: %v", err)
					break
				}
			}
		}()
	})

	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		connections.Add(1)

		notify := r.Context().Done()

		for {
			select {
			case <-notify:
				connections.Add(-1)
				return
			case <-time.After(1 * time.Second):
				_, err := w.Write([]byte("data: hello\n\n"))
				if err != nil {
					log.Printf("write error: %v", err)
					return
				}
				flusher.Flush()
			}
		}
	})

	err := http.ListenAndServe(":6969", mux)

	if err != nil {
		log.Fatal(err)
	}

}
