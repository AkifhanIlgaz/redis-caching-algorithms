package main

import (
	"context"
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

	fifo := cache.NewFIFO(ctx, client, 3, "fifocache")

	users := []cache.User{
		{Name: "Alice", Age: 30, Id: "1"},
		{Name: "Bob", Age: 25, Id: "2"},
		{Name: "Charlie", Age: 35, Id: "3"},
	}

	for _, user := range users {
		fifo.Put(user)
	}

}
