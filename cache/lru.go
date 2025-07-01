package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// LRUCache represents a Least Recently Used (LRU) cache implemented with Redis.
// It uses a Redis sorted set to maintain the order of items by their last access time.
type LRUCache struct {
	ctx       context.Context
	client    *redis.Client
	keyPrefix string
	capacity  int
}

// NewLRU creates a new LRUCache with the given context, Redis client, capacity, and key prefix.
func NewLRU(ctx context.Context, client *redis.Client, capacity int, keyPrefix string) LRUCache {
	log.Println("Creating new LRU cache with capacity:", capacity)
	return LRUCache{
		ctx:       ctx,
		client:    client,
		capacity:  capacity,
		keyPrefix: keyPrefix,
	}
}

// MakeRequest simulates a request for a user by their ID.
// It first tries to get the user from the cache. If the user is not in the cache (a cache miss),
// it fetches the user from the database, adds them to the cache, and then returns the user.
// If the user is found in the cache (a cache hit), it returns the user directly.
func (c *LRUCache) MakeRequest(id string) User {
	log.Printf("Request received for user ID: %s", id)
	user, err := c.Get(id)
	if err != nil {
		log.Printf("Cache miss for user ID: %s. Fetching from database.", id)
		dbUser := getUserFromDb(id)
		if err := c.Set(dbUser); err != nil {
			log.Printf("Failed to write user ID: %s to cache: %v", id, err)
		}
		return dbUser
	}

	log.Printf("Cache hit for user ID: %s.", id)
	return user
}

// Get retrieves a user from the cache by their ID.
// If the user is found, it updates their recency and returns the user.
func (c *LRUCache) Get(id string) (User, error) {
	cacheKey := c.generateKey(userPrefix, id)
	log.Printf("Attempting to get user with cache key: %s", cacheKey)

	data, err := c.client.Get(c.ctx, cacheKey).Result()
	if err != nil {
		log.Printf("Error getting user with cache key: %s from Redis: %v", cacheKey, err)
		return User{}, err
	}

	log.Printf("Successfully retrieved user with cache key: %s. Updating recency.", cacheKey)
	if err := c.UpdateRecency(id); err != nil {
		log.Printf("Failed to update recency for user ID: %s: %v", id, err)
		return User{}, err
	}

	var user User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		log.Printf("Error unmarshalling user data for cache key: %s: %v", cacheKey, err)
		return User{}, err
	}

	return user, nil
}

// Set adds a user to the cache.
// If the cache is full, it removes the oldest item before adding the new one.
func (c *LRUCache) Set(user User) error {
	log.Printf("Attempting to set user with ID: %s to cache.", user.Id)
	currentSize := c.CacheSize()
	if currentSize >= c.capacity {
		log.Printf("Cache is full (size: %d, capacity: %d). Removing oldest item.", currentSize, c.capacity)
		if err := c.RemoveOldest(); err != nil {
			log.Printf("Failed to remove oldest item from cache: %v", err)
			return err
		}
	}

	return c.AddKey(user)
}

// Delete removes a key from the cache.
func (c *LRUCache) Delete(key string) error {
	log.Printf("Deleting key: %s from cache", key)
	return c.client.Del(c.ctx, key).Err()
}

// CacheSize returns the current number of items in the cache.
func (c *LRUCache) CacheSize() int {
	key := c.generateKey(cacheKeyPrefix)
	log.Printf("Getting cache size for key: %s", key)

	size, err := c.client.ZCard(c.ctx, key).Result()
	if err != nil {
		log.Printf("Error getting cache size for key: %s. Error: %v", key, err)
		return 0
	}
	log.Printf("Cache size for key: %s is: %d", key, size)
	return int(size)
}

// AddKey adds a new user to the cache. It adds the user's data to a Redis key
// and adds the key to the sorted set for LRU tracking.
func (c *LRUCache) AddKey(user User) error {
	listKey := c.generateKey(cacheKeyPrefix)
	cacheKey := c.generateKey(userPrefix, user.Id)
	log.Printf("Adding key: %s to list: %s", cacheKey, listKey)

	if err := c.client.ZAdd(c.ctx, listKey, redis.Z{
		Member: cacheKey,
		Score:  float64(time.Now().Unix()),
	}).Err(); err != nil {
		log.Printf("Error adding key: %s to sorted set: %s: %v", cacheKey, listKey, err)
		return err
	}

	b, err := json.Marshal(&user)
	if err != nil {
		log.Printf("Error marshalling user data for ID: %s: %v", user.Id, err)
		return err
	}

	log.Printf("Setting value for key: %s", cacheKey)
	return c.client.Set(c.ctx, cacheKey, b, 0).Err()
}

// UpdateRecency updates the access time of a user in the cache, marking them as recently used.
func (c *LRUCache) UpdateRecency(id string) error {
	listKey := c.generateKey(cacheKeyPrefix)
	cacheKey := c.generateKey(userPrefix, id)
	log.Printf("Updating recency for key: %s in list: %s", cacheKey, listKey)

	if err := c.client.ZAdd(c.ctx, listKey, redis.Z{
		Member: cacheKey,
		Score:  float64(time.Now().Unix()),
	}).Err(); err != nil {
		log.Printf("Error updating recency for key: %s: %v", cacheKey, err)
		return err
	}

	return nil
}

// RemoveOldest removes the least recently used item from the cache.
func (c *LRUCache) RemoveOldest() error {
	listKey := c.generateKey(cacheKeyPrefix)
	log.Printf("Removing oldest item from list: %s", listKey)

	removed, err := c.client.ZPopMin(c.ctx, listKey, 1).Result()
	if err != nil {
		log.Printf("Error removing oldest item from sorted set: %s: %v", listKey, err)
		return err
	}

	if len(removed) == 0 {
		log.Println("No items to remove from cache.")
		return fmt.Errorf("no items to remove from cache")
	}

	removedMember := removed[0].Member.(string)
	log.Printf("Popped oldest member: %s", removedMember)

	return c.Delete(removedMember)
}

// generateKey creates a Redis key by joining the key prefix and other key parts with a colon.
func (c *LRUCache) generateKey(keys ...string) string {
	allKeys := []string{c.keyPrefix}
	allKeys = append(allKeys, keys...)

	return strings.Join(allKeys, ":")
}
