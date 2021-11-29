package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// Digits are chosen such that we cannot easily confuse them (e.g.: 0 and O)
	Universe = "23456789CFGHJMPQRVWX" // Pr(20, 5) = 3,200,000
	RandSeed = 42
)

var (
	Rand = rand.New(rand.NewSource(RandSeed))
)

func randomRoomName() string {
	var roomId strings.Builder
	max := len(Universe)
	for k := 0; k < 5; k++ {
		randIdx := Rand.Intn(max)
		roomId.WriteByte(Universe[randIdx])
	}
	return roomId.String()
}

type RoomManager interface {
	JoinRoom(room string) error
	CreateRoom() (string, error)
}

type RedisState struct {
	Rdb *redis.Client
	ctx context.Context
}

func NewRedisState(ctx context.Context, rdb *redis.Client) *RedisState {
	return &RedisState{
		Rdb: rdb,
		ctx: ctx,
	}
}

func NewRedisStateWithDefaults() *RedisState {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return NewRedisState(ctx, rdb)
}

type InternalState map[string]string
type LocalState struct {
	KVStore InternalState
}

func NewLocalState() *LocalState {
	return &LocalState{
		KVStore: make(InternalState),
	}
}

func (s *LocalState) JoinRoom(room string) error {
	_, found := s.KVStore["room:"+room]
	fmt.Println(found)
	if !found {
		return errors.New("No such room.")
	}
	return nil
}

func (s *LocalState) CreateRoom() (string, error) {
	room := randomRoomName()
	s.KVStore["room:"+room] = strconv.Itoa(int(time.Now().UTC().Unix()))
	return room, nil
}

func (s *RedisState) JoinRoom(room string) error {
	err := s.Rdb.Get(s.ctx, "room:"+room).Err()
	return err
}

func (s *RedisState) CreateRoom() (string, error) {
	room := randomRoomName()
	val := time.Now().UTC().Unix()
	err := s.Rdb.Set(s.ctx, "room:"+room, val, 0).Err()
	return room, err
}
