package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AkifhanIlgaz/redis-caching-algorithms/cache"
	"github.com/redis/go-redis/v9"
)

const connectionString string = "redis://@localhost:6379/0"

func main() {
	opt, err := redis.ParseURL(connectionString)
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Clear the database before starting
	client.FlushDB(ctx)
	// Create a new FIFO cache with a capacity of 3
	fifoCache := cache.NewLFU(ctx, client, 3, "lfu_cache")

	fmt.Println("-------------------------------------------------------- Press any key to continue")
	fmt.Scanln()
	// Request a user that is not in the cache
	user1 := fifoCache.MakeRequest("1")
	fmt.Printf("Got user: %v\n", user1)
	fmt.Println("--------------------------------------------------------")
	fmt.Scanln()
	// Request the same user again, this time it should be a cache hit
	user1_cached := fifoCache.MakeRequest("1")
	fmt.Printf("Got user from cache: %v\n", user1_cached)
	fmt.Println("--------------------------------------------------------")
	fmt.Scanln()
	// Add two more users to fill the cache
	fifoCache.MakeRequest("2")
	fmt.Println("--------------------------------------------------------")
	fmt.Scanln()
	fifoCache.MakeRequest("3")
	fmt.Println("--------------------------------------------------------")
	fmt.Scanln()
	// Add one more user, this should evict the first user (user1)
	fifoCache.MakeRequest("4")
	fmt.Println("--------------------------------------------------------")
	fmt.Scanln()
	// Request user1 again, this should be a cache miss and fetched from the db
	user1_after_eviction := fifoCache.MakeRequest("1")
	fmt.Printf("Got user after eviction: %v\n", user1_after_eviction)
	fmt.Println("--------------------------------------------------------")
	fmt.Scanln()

}
