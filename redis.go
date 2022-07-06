package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	redis "github.com/go-redis/redis/v9"
)

const (
	REDIS_ADDR               = "localhost:6379"
	EXPIRY_TIME              = time.Second * 30
	SLEEP_TIME               = time.Second * 15
	CONV_REDIS_MATCH_PATTERN = "{*}"
)

var (
	gRedisClient  *redis.Client
	gRedisContext = context.Background()
)

func init() {
	gRedisClient = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDR,
		Password: "",
		DB:       0,
	})
}

func SendConvListToRedis(convs []Conv) {
	for _, conv := range convs {
		if bs, err := json.Marshal(conv); err != nil {
			log.Println("Redis, marshal conv error: ", err)
		} else {
			gRedisClient.Set(gRedisContext, string(bs), gServerId, EXPIRY_TIME)
		}
	}
}

func GetConvListFromRedis() (convs []Conv) {
	cursor := uint64(0)

	for {
		var keys []string
		var err error
		keys, cursor, err = gRedisClient.Scan(gRedisContext, cursor, CONV_REDIS_MATCH_PATTERN, 0).Result()

		if err != nil {
			log.Println("Get ConvList from redis error: ", err)
			return
		}

		for _, key := range keys {
			if conv, err := NewRemoteConvFromJSON(key); err != nil {
				log.Printf("Marshal JSON RemoveConv error: %s\n", err)
			} else {
				convs = append(convs, conv)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return
}

func StartSendConvListToRedis() {
	for {
		before := time.Now()

		convs := gConvs.Values()
		SendConvListToRedis(convs)

		log.Printf("%s sent ConvList to redis\n", gServerId)

		after := time.Now()

		elapsed := after.Sub(before)
		if elapsed < SLEEP_TIME {
			time.Sleep(SLEEP_TIME - elapsed)
		}
	}
}
