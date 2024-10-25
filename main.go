package main

import (
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	err := http.ListenAndServe(":6969", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Printf("upgrade error: %v", err)
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					log.Printf("read error: %v", err)
				}
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					log.Printf("write error: %v", err)
				}
			}
		}()
	}))

	if err != nil {
		log.Fatal(err)
	}
}
