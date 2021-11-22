package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestJoinWithParams(t *testing.T) {
	wrapPlay := func(w http.ResponseWriter, r *http.Request) {
		play([]byte("seed_test"), w, r)
	}
	ts := httptest.NewServer(http.HandlerFunc(wrapPlay))
	defer ts.Close()
	wsURL := strings.Replace(ts.URL, "http://", "ws://", 1) + "?username=player&room=ROOM"
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Errorf("WS Error: %q", err.Error())
		return
	}
	defer c.Close()
}

func TestJoinWithoutParams(t *testing.T) {
	wrapPlay := func(w http.ResponseWriter, r *http.Request) {
		play([]byte("seed_test"), w, r)
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
	seed := "seed_for_test"
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	defer server.Close()
	champ, err := NewChampignon(seed, server)
	go champ.Serve()
	t.Log(champ)
	t.Log(err)
}
