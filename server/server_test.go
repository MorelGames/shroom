package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

const (
	seedTest = "SEED_TEST"
)

func TestJoinWithoutParams(t *testing.T) {
	s, err := miniredis.Run() // Fake local Redis server
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	state := NewRedisState(
		context.Background(),
		redis.NewClient(&redis.Options{
			Addr:     s.Addr(),
			Password: "",
			DB:       0,
		}),
	)
	wrapPlay := func(w http.ResponseWriter, r *http.Request) {
		play([]byte(seedTest), state, w, r)
	}
	ts := httptest.NewServer(http.HandlerFunc(wrapPlay))
	defer ts.Close()
	wsURL := strings.Replace(ts.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		defer c.Close()
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
	if err != nil {
		t.Error(err.Error())
		return
	}
	go champ.Serve()
}

func joinCreateRoom(state *RedisState, t *testing.T) {
	err := state.JoinRoom("ROOM", "USER")
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
	err = state.JoinRoom(room, "USER")
	if err != nil {
		t.Errorf("Room %s must exists at this point", room)
	}
}

func TestJoinRoomCreateRoomLocal(t *testing.T) {
	s, err := miniredis.Run() // Fake local Redis server
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer s.Close()
	state := NewRedisState(
		context.Background(),
		redis.NewClient(&redis.Options{
			Addr:     s.Addr(),
			Password: "",
			DB:       0,
		}),
	)
	joinCreateRoom(state, t)
}

func TestRoomInfo(t *testing.T) {
	s, err := miniredis.Run() // Fake local Redis server
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer s.Close()
	state := NewRedisState(
		context.Background(),
		redis.NewClient(&redis.Options{
			Addr:     s.Addr(),
			Password: "",
			DB:       0,
		}),
	)
	username := "user42"
	room, _ := state.CreateRoom()
	roomInfo, _ := state.RoomInfo(room)

	want := []string{} // Room has been created, but there is no player yet
	got := roomInfo.Players
	if !reflect.DeepEqual(want, got) {
		t.Errorf("players: got %v, want %v", got, want)
	}

	// Player joins the room
	state.JoinRoom(room, username)
	roomInfo, _ = state.RoomInfo(room) // Refetch infos

	want = []string{username} // Now there is 1 player
	got = roomInfo.Players
	if !reflect.DeepEqual(want, got) {
		t.Errorf("players: got %v, want %v", got, want)
	}

	// Same username: error and no effect
	err = state.JoinRoom(room, username)
	if err == nil {
		t.Error("Expected error: same username")
		return
	}

	// Another username
	state.JoinRoom(room, username+"other")
	roomInfo, _ = state.RoomInfo(room)

	want1 := []string{"user42", "user42other"}
	want2 := []string{"user42other", "user42"}
	got = roomInfo.Players
	if !reflect.DeepEqual(want1, got) && !reflect.DeepEqual(want2, got) {
		t.Errorf("players: got %v, want %v", got, want)
	}
}

func TestRoomInfoWithUTF8(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	defer s.Close()
	state := NewRedisState(
		context.Background(),
		redis.NewClient(&redis.Options{
			Addr:     s.Addr(),
			Password: "",
			DB:       0,
		}),
	)
	username := "猫ちゃん🐱"
	room, _ := state.CreateRoom()
	state.JoinRoom(room, username)
	roomInfo, _ := state.RoomInfo(room)

	want := []string{username}
	got := roomInfo.Players
	if !reflect.DeepEqual(want, got) {
		t.Errorf("players: got %v, want %v", got, want)
	}
	t.Log(roomInfo.Players)
}
