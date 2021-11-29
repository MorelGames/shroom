package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

const (
	seedTest = "SEED_TEST"
	roomTest = "R00M"
)

func TestJoinWithoutParams(t *testing.T) {
	state := NewLocalState()
	wrapPlay := func(w http.ResponseWriter, r *http.Request) {
		play([]byte(seedTest), state, w, r)
	}
	ts := httptest.NewServer(http.HandlerFunc(wrapPlay))
	defer ts.Close()
	wsURL := strings.Replace(ts.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		defer c.Close()
		t.Errorf("Expected error here")
		t.Log(c)
	}
}

func TestNewChampignon(t *testing.T) {
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	defer server.Close()
	champ, err := NewChampignon(seedTest, server)
	go champ.Serve()
	t.Log(champ)
	t.Log(err)
}

func TestRoomManagerIface(t *testing.T) {
	var _ RoomManager
}

func joinCreateRoom(state RoomManager, t *testing.T) {
	err := state.JoinRoom("ROOM")
	if err == nil {
		t.Errorf("Expected error here, because the room cannot exist yet!")
		return
	}
	room, err := state.CreateRoom()
	if err != nil {
		t.Errorf("Got error: %q", err)
		return
	}
	if len(room) != 5 {
		t.Errorf("Invalid room name: %s", room)
		return
	}
	err = state.JoinRoom(room)
	if err != nil {
		t.Errorf("Room %s must exists at this point", room)
	}
}

func TestJoinRoomCreateRoomLocal(t *testing.T) {
	state := NewLocalState()
	joinCreateRoom(state, t)
}
