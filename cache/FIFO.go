package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

const userPrefix = "user"
const cacheKeyPrefix = "cache_key"

// FIFOCache represents a LRU cache implemented with linked list in Redis.
type FIFOCache struct {
	ctx       context.Context
	client    *redis.Client
	keyPrefix string
	capacity  int
}

// NewFIFO creates a new FIFOCache.
func NewFIFO(ctx context.Context, client *redis.Client, capacity int, keyPrefix string) FIFOCache {
	log.Println("Creating new FIFO cache")
	return FIFOCache{
		ctx:       ctx,
		client:    client,
		capacity:  capacity,
		keyPrefix: keyPrefix,
	}
}

// MakeRequest retrieves a user. It first tries to get the user from the cache.
// If the user is not in the cache, it gets the user from the database and adds it to the cache.
func (c *FIFOCache) MakeRequest(id string) User {
	log.Printf("Making request for user with id: %s", id)
	user, err := c.Get(id)
	if err != nil {
		log.Printf("Cache miss for user with id: %s. Getting from DB.", id)
		dbUser := getUserFromDb(id)
		if err := c.Set(dbUser); err != nil {
			log.Printf("Cannot write to cache")
		}
		return dbUser
	}

	log.Printf("Cache hit for user with id: %s.", id)
	return user
}

// Get retrieves a user from the cache.
func (c *FIFOCache) Get(id string) (User, error) {
	cacheKey := c.generateKey(userPrefix, id)

	log.Printf("Getting user with key: %s from cache", cacheKey)
	data, err := c.client.Get(c.ctx, cacheKey).Result()
	if err != nil {
		return User{}, err
	}

	var user User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// Set adds a user to the cache. If the cache is full, it removes the oldest item before adding the new one.
func (c *FIFOCache) Set(user User) error {
	log.Printf("Setting user with id: %s to cache", user.Id)
	if c.CacheSize() == c.capacity {
		log.Println("Cache is full. Removing oldest item.")
		err := c.RemoveOldest()
		if err != nil {
			return err
		}
	}

	return c.AddKey(user)
}

// Delete removes a key from the cache.
func (c *FIFOCache) Delete(key string) error {
	log.Printf("Deleting key: %s from cache", key)
	return c.client.Del(c.ctx, key).Err()
}

// CacheSize returns the current number of items in the cache.
func (c *FIFOCache) CacheSize() int {
	key := c.generateKey(cacheKeyPrefix)
	log.Printf("Getting cache size for key: %s", key)

	size, err := c.client.LLen(c.ctx, key).Result()
	if err != nil {
		log.Printf("Error getting cache size for key: %s. Error: %v", key, err)
		return 0
	}
	log.Printf("Cache size for key: %s is: %d", key, size)
	return int(size)
}

// AddKey adds a new key to the cache.
func (c *FIFOCache) AddKey(user User) error {
	listKey := c.generateKey(cacheKeyPrefix)
	cacheKey := c.generateKey(userPrefix, user.Id)
	log.Printf("Adding key: %s to list: %s", cacheKey, listKey)

	if err := c.client.RPush(c.ctx, listKey, cacheKey).Err(); err != nil {
		return err
	}

	b, err := json.Marshal(&user)
	if err != nil {
		return err
	}

	log.Printf("Setting value for key: %s", cacheKey)
	return c.client.Set(c.ctx, cacheKey, b, 0).Err()
}

// RemoveOldest removes the oldest item from the cache.
func (c *FIFOCache) RemoveOldest() error {
	listKey := c.generateKey(cacheKeyPrefix)
	log.Printf("Removing oldest item from list: %s", listKey)
	removedKey, err := c.client.LPop(c.ctx, listKey).Result()
	if err != nil {
		return err
	}

	log.Printf("Removed key: %s", removedKey)
	return c.Delete(removedKey)
}

// generateKey creates a Redis key by joining the given parts with a colon.
func (c *FIFOCache) generateKey(keys ...string) string {
	allKeys := []string{c.keyPrefix}
	allKeys = append(allKeys, keys...)

	return strings.Join(allKeys, ":")
}
