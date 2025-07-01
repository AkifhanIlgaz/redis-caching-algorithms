package cache

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/redis/go-redis/v9"
)

const userPrefix = "user"
const keyListPrefix = "key"

type FIFOCache struct {
	ctx       context.Context
	client    *redis.Client
	keyPrefix string
	capacity  int
}

func NewFIFO(ctx context.Context, client *redis.Client, capacity int, keyPrefix string) FIFOCache {
	return FIFOCache{
		ctx:       ctx,
		client:    client,
		capacity:  capacity,
		keyPrefix: keyPrefix,
	}
}

func (c *FIFOCache) MakeRequest(id string) *User {
	user := c.GetFromCache(id)

	// Cache miss
	if user == nil {
		user := getUserFromDb(id)
		c.Put(*user)
		return user
	}

	// Cache hit
	return user
}

func (c *FIFOCache) GetFromCache(id string) *User {
	key := c.GenerateKey(userPrefix, id)

	// Cache hit
	data, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		return nil
	}

	var user User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil
	}

	return &user
}

func (c *FIFOCache) Put(user User) {
	if c.Size() == c.capacity {
		removedKey := c.RemoveFromList()
		c.RemoveFromCache(removedKey)
	}

	c.AddToList(user.Id)
	b, err := json.Marshal(&user)
	if err != nil {
		panic(err)
	}
	key := c.GenerateKey(userPrefix, user.Id)

	err = c.client.Set(c.ctx, key, b, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (c *FIFOCache) Size() int {
	key := c.GenerateKey(keyListPrefix)

	size, err := c.client.LLen(c.ctx, key).Result()
	if err != nil {
		return 0
	}
	return int(size)
}

func (c *FIFOCache) RemoveFromList() string {
	key := c.GenerateKey(keyListPrefix)
	id, _ := c.client.LPop(c.ctx, key).Result()
	return id
}

func (c *FIFOCache) RemoveFromCache(key string) {
	c.client.Del(c.ctx, key)
}

func (c *FIFOCache) AddToList(id string) {
	key := c.GenerateKey(keyListPrefix)
	valueKey := c.GenerateKey(userPrefix, id)
	c.client.RPush(c.ctx, key, valueKey)
}

func (c *FIFOCache) GenerateKey(keys ...string) string {
	allKeys := []string{c.keyPrefix}
	allKeys = append(allKeys, keys...)

	return strings.Join(allKeys, ":")
}
