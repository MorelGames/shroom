package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJoinWithParams(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(play))
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
	ts := httptest.NewServer(http.HandlerFunc(play))
	defer ts.Close()
	wsURL := strings.Replace(ts.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		defer c.Close()
		t.Errorf("Expected error here")
		t.Log(c)
	}
}
