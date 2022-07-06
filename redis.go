package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v9"
)

const (
	REDIS_ADDR   = "0.0.0.0:6379"
	CHAT_CHANNEL = "chat"
	EXPIRY_TIME  = time.Second * 30
	SLEEP_TIME   = time.Second * 15

	// since conv saved in redis in the form of key: json string - value: server id
	// We can use below pattern to match every conv
	CONV_REDIS_MATCH_PATTERN = "{*}"
)

var (
	gRedisClient      *RedisClient = nil
	gRedisClientMutex sync.Mutex
)

type RedisClient struct {
	client *redis.Client
	pubsub *redis.PubSub
	ctx    context.Context
}

func GetRedisClient() *RedisClient {
	if gRedisClient == nil {
		gRedisClientMutex.Lock()
		defer gRedisClientMutex.Unlock()

		if gRedisClient == nil {
			ctx := context.Background()
			client := redis.NewClient(&redis.Options{
				Addr:     REDIS_ADDR,
				Password: "",
				DB:       0,
			})
			pubsub := client.Subscribe(ctx, CHAT_CHANNEL)
			if _, err := pubsub.Receive(ctx); err != nil {
				panic(err)
			}

			gRedisClient = &RedisClient{
				ctx:    ctx,
				client: client,
				pubsub: pubsub,
			}
		}
	}
	return gRedisClient
}

func (redis *RedisClient) SendMessage(msg ServerRequestMessage) {
	log.Println("Sent a message to redis", msg)
	go func() {
		if bs, err := json.Marshal(msg); err != nil {
			log.Println("Redis, marshal conv error: ", err)
		} else {
			if _, err := redis.client.Publish(redis.ctx, CHAT_CHANNEL, string(bs)).Result(); err != nil {
				log.Println(err)
			}
		}
	}()
}

func (redis *RedisClient) SendConvList(convs []Conv) {
	for _, conv := range convs {
		if bs, err := json.Marshal(conv); err != nil {
			log.Println("Redis, marshal conv error: ", err)
		} else {
			redis.client.Set(redis.ctx, string(bs), gServerId, EXPIRY_TIME)
		}
	}
}

func (redis *RedisClient) GetConvList() (convs []Conv) {
	cursor := uint64(0)

	for {
		var keys []string
		var err error
		keys, cursor, err = redis.client.Scan(redis.ctx, cursor, CONV_REDIS_MATCH_PATTERN, 0).Result()

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

		// if visited all keys that match conversation id pattern
		if cursor == 0 {
			break
		}
	}

	return
}

func (redis *RedisClient) StartSendConvList() {
	for {
		before := time.Now()

		convs := gConvs.Values()
		redis.SendConvList(convs)

		log.Printf("%s sent ConvList to redis\n", gServerId)

		after := time.Now()

		elapsed := after.Sub(before)
		if elapsed < SLEEP_TIME {
			time.Sleep(SLEEP_TIME - elapsed)
		}
	}
}

func (redis *RedisClient) StartListening() {
	for msg := range redis.pubsub.Channel() {
		redis.processRequest(msg.Payload)
	}
}

func (redis *RedisClient) processRequest(payload string) {
	var req ServerRequestMessage
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		log.Printf("redis processRequest, Unmarshal to %T error: %s\n", req, err)
		return
	}

	log.Println(req)

	if req.SenderServerId == gServerId {
		return
	}

	switch req.Request {
	case TEXT_OTHER_SERVER_ACTION:
		if msg, ok := marshalJSON[TextMessage](req.Data); ok {
			gConvs.RApplyToOne(
				func(_ string, conv Conv) bool { return conv.Id() == msg.ReceiverId },
				func(_ string, conv Conv) { conv.DeliverTextMessage(msg) },
			)
		}
	case CLIENT_CONNECTED_ACTION:
		if client, ok := marshalJSON[ClientConnectedMessage](req.Data); ok {
			conv := RemoteConv{
				ID:       client.Id,
				ServerID: client.ServerId,
				Type:     PEER_TYPE,
			}
			gConvs.Store(conv.ID, conv)
		}
	case CLIENT_DISCONNECTED_ACTION:
		if clientId, ok := req.Data.(string); !ok {
			log.Printf("redis processRequest, cannot parse clientId in %s\n", CLIENT_DISCONNECTED_ACTION)
		} else {
			gRemoveClient(clientId)
		}
	case GROUP_CREATED_ACTION:
		if client, ok := marshalJSON[ClientConnectedMessage](req.Data); ok {
			conv := RemoteConv{
				ID:       client.Id,
				ServerID: client.ServerId,
				Type:     GROUP_TYPE,
			}
			gConvs.Store(conv.ID, conv)
		}
	}
}
