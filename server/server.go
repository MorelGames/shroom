package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{}
var appSecret = []byte("this_is_secret")
var creationDate = time.Date(2021, 11, 16, 17, 04, 22, 324243, time.UTC)
var creationDateMs = creationDate.UnixNano() / 1e6 // Use UnixMilli() with Go 1.17
var interval = 15000                               // msecs

type SafeConn struct {
	mu sync.Mutex
	c  *websocket.Conn
}

func (sc *SafeConn) WriteJSON(v interface{}) error {
	sc.mu.Lock()
	err := sc.c.WriteJSON(v)
	sc.mu.Unlock()
	return err
}

func main() {
	if time.Now().Before(creationDate) {
		fmt.Println("creationDate", creationDate)
		fmt.Println("Now", time.Now())
		panic("creationDate < Now")
	}
	port := os.Args[1]
	http.HandleFunc("/play", play) // WS handler
	http.ListenAndServe("[::1]:"+port, nil)
}

func play(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if len(q["room"]) == 0 || len(q["username"]) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	sc := &SafeConn{c: c}

	username := q["username"][0]
	room := q["room"][0]

	// goroutine that sends new questions when in due time
	go func() {
		fmt.Println("Room creation date:", creationDate)
		roomSeed := append(appSecret, []byte(creationDate.String())...)
		for {
			now := time.Now()
			nowMs := now.UnixNano() / 1e6 // Use UnixMilli() with Go 1.17
			delta := int(nowMs - creationDateMs)
			tItlv := delta - (delta % interval)
			remainingMs := tItlv + interval - delta

			rg := Rg(roomSeed, uint32(tItlv)) % 500

			msg := struct {
				Username string    `json:"username"`
				Question int       `json:"question"`
				Exp      time.Time `json:"exp"`
				Room     string    `json:"room"`
			}{
				username,
				rg,
				now.Add(time.Duration(remainingMs+2000) * time.Millisecond), // Add 2 secs margin
				room,
			}
			err := sc.WriteJSON(&msg)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			time.Sleep(time.Duration(remainingMs) * time.Millisecond)
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err)
			break
		}
		msg := struct {
			Msg string
			Now time.Time
		}{
			string(message),
			time.Now().UTC(),
		}
		sc.WriteJSON(&msg)
	}
}
