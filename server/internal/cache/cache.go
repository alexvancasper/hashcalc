package cache

import (
	"context"
	"encoding/json"
	"log"
	"module13/internal/entities"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type record struct {
	entity entities.User
}

type Cache struct {
	db *redis.Client
}

func (r *record) MarshalBinary() (data []byte, err error) {
	bytes, err := json.Marshal(r.entity)
	return bytes, err
}

func (r *record) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, &r.entity)
	return err
}

func (c *Cache) Set(key string, e entities.User, time time.Duration) {
	r := record{entity: e}
	data, err := r.MarshalBinary()
	if err != nil {
		log.Printf("[Cache Set] Marshalling error %s", err)
		return
	}
	err = c.db.Set(context.TODO(), key, data, time).Err()
	if err != nil {
		log.Printf("[Cache Set] Entity added failed %s", err)
	}
}

func (c *Cache) Get(key string) (entities.User, bool) {
	var buf []byte
	err := c.db.Get(context.TODO(), key).Scan(&buf)

	if err == redis.Nil {
		log.Printf("key does not exist %s", key)
		return entities.User{}, false
	} else if err != nil {
		log.Printf("[Cache Get] failed to fetch data")
		return entities.User{}, false
	}
	r := record{}
	err = r.UnmarshalBinary(buf)
	if err != nil {
		log.Printf("[Cache GET] Unmarshal error %s", err)
	}

	return r.entity, true

}

func NewCache() *Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: os.Getenv("REDIS_PASS"), // no password set
		DB:       0,                       // use default DB
	})
	status := rdb.Ping(context.TODO())
	if status.Err() != nil {
		panic(status.Err())
	}
	result := rdb.Info(context.TODO())
	log.Printf("%+v\n", result)
	return &Cache{db: rdb}
}
