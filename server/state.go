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
)

func randomRoomName() string {
	var roomId strings.Builder
	r := rand.New(rand.NewSource(time.Now().Unix()))
	max := len(Universe)
	for k := 0; k < 5; k++ {
		randIdx := r.Intn(max)
		roomId.WriteByte(Universe[randIdx])
	}
	return roomId.String()
}

type RoomInfo struct {
	Room    string
	Created time.Time
	Game    int
	Players []string
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
	KHStore map[string]InternalState
}

func NewLocalState() *LocalState {
	return &LocalState{
		KVStore: make(InternalState),
		KHStore: make(map[string]InternalState),
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
	val := strconv.FormatInt(time.Now().Unix(), 10)
	s.KVStore["room:"+room] = val
	return room, nil
}

func (s *RedisState) JoinRoom(room string, username string) error {
	val, _ := s.Rdb.HGetAll(s.ctx, "room:"+room).Result()
	if len(val) == 0 {
		return errors.New("No such room: " + room)
	}
	added, _ := s.Rdb.SAdd(s.ctx, "room:"+room+":players", username).Result()
	if added == 0 { // Redis added 0 user, i.e., this username is already here
		return errors.New("Username taken: " + username)
	}
	return nil
}

func (s *RedisState) CreateRoom() (string, error) {
	room := randomRoomName()
	val := strconv.FormatInt(time.Now().Unix(), 10)
	err := s.Rdb.HSet(s.ctx, "room:"+room, []string{
		"created", val,
		"game", "",
	}).Err()
	return room, err
}

func (s *RedisState) RoomInfo(room string) (*RoomInfo, error) {
	kv, _ := s.Rdb.HGetAll(s.ctx, "room:"+room).Result()
	if len(kv) == 0 {
		return &RoomInfo{}, errors.New("No such room: " + room)
	}
	players, err := s.Rdb.SMembers(s.ctx, "room:"+room+":players").Result()
	if err != nil {
		return &RoomInfo{}, err
	}
	val, _ := kv["created"]
	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return &RoomInfo{}, err
	}
	created := time.Unix(timestamp, 0)
	val, _ = kv["game"]
	if err != nil {
		return &RoomInfo{}, err
	}
	game, _ := strconv.Atoi(val)
	roomInfo := &RoomInfo{
		Room:    room,
		Game:    game,
		Players: players,
		Created: created,
	}
	return roomInfo, nil
}

func (s *RedisState) NewGame(room string) error {
	info, err := s.RoomInfo(room)
	if err != nil {
		return err
	}
	if len(info.Players) == 0 {
		return errors.New("Cannot run a new game without any player.")
	}
	game := strconv.Itoa(int(time.Now().Unix()))
	err = s.Rdb.HSet(s.ctx, "room:"+room, []string{
		"game", game,
	}).Err()
	if err != nil {
		return err
	}
	return nil
}
