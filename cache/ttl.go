package cache

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// TTLCache represents a Time To Live (TTL) cache implemented with Redis.
// It sets an expiration time for each key, and Redis automatically handles the eviction
// of expired keys. This cache is effective for data that becomes stale after a certain period.
type TTLCache struct {
	ctx        context.Context
	client     *redis.Client
	expiration time.Duration
	keyPrefix  string
}

// NewTTL initializes and returns a new TTLCache.
//
// Parameters:
//   - ctx: The context for Redis operations.
//   - client: The Redis client instance.
//   - expiration: The duration for which each cache entry should be valid.
//   - keyPrefix: A prefix for all cache keys to avoid collisions.
//
// Returns:
//   A new instance of TTLCache.
func NewTTL(ctx context.Context, client *redis.Client, expiration time.Duration, keyPrefix string) TTLCache {
	return TTLCache{
		ctx:        ctx,
		client:     client,
		keyPrefix:  keyPrefix,
		expiration: expiration,
	}
}

// MakeRequest handles a request for a user by their ID, using the TTL cache.
// It first attempts to retrieve the user from the cache. If the user is not found (a cache miss),
// it fetches the user from the database, stores the new user in the cache with a defined TTL,
// and then returns the user. If the user is found in the cache (a cache hit), it returns the user directly.
// This method is ideal for scenarios where data should be cached for a specific duration.
//
// Parameters:
//   - id: The ID of the user to request.
//
// Returns:
//   The requested User object.
func (c *TTLCache) MakeRequest(id string) User {
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

// Get retrieves a user from the cache by their ID. It fetches the value from Redis
// and unmarshals it into a User object. This method is a straightforward key-value lookup
// and does not involve any TTL management, as Redis handles expiration automatically.
//
// Parameters:
//   - id: The ID of the user to retrieve.
//
// Returns:
//   The User object and an error if the user is not found or if unmarshalling fails.
func (c *TTLCache) Get(id string) (User, error) {
	cacheKey := c.generateKey(userPrefix, id)
	log.Printf("Attempting to get user with cache key: %s", cacheKey)

	data, err := c.client.Get(c.ctx, cacheKey).Result()
	if err != nil {
		log.Printf("Error getting user with cache key: %s from Redis: %v", cacheKey, err)
		return User{}, err
	}

	var user User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		log.Printf("Error unmarshalling user data for cache key: %s: %v", cacheKey, err)
		return User{}, err
	}

	return user, nil
}

// Set adds a user to the cache with the configured TTL. It marshals the User object
// to JSON and stores it in Redis. The key is generated using the user's ID,
// and the entry is set to expire after the predefined duration.
//
// Parameters:
//   - user: The User object to store in the cache.
//
// Returns:
//   An error if marshalling or the Redis SET operation fails.
func (c *TTLCache) Set(user User) error {
	cacheKey := c.generateKey(userPrefix, user.Id)

	b, err := json.Marshal(&user)
	if err != nil {
		log.Printf("Error marshalling user data for ID: %s: %v", user.Id, err)
		return err
	}

	log.Printf("Setting value for key: %s", cacheKey)
	return c.client.Set(c.ctx, cacheKey, b, c.expiration).Err()
}

// generateKey constructs a Redis key by joining the configured key prefix
// with the provided key parts, separated by colons. This ensures consistent
// and unique key naming within the cache.
//
// Parameters:
//   - keys: A variadic slice of strings that make up the key parts.
//
// Returns:
//   A single string representing the full Redis key.
func (c *TTLCache) generateKey(keys ...string) string {
	allKeys := []string{c.keyPrefix}
	allKeys = append(allKeys, keys...)

	return strings.Join(allKeys, ":")
}
