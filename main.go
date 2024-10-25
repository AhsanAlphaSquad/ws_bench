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

	err := http.ListenAndServe(":6969", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	if err != nil {
		log.Fatal(err)
	}
}
