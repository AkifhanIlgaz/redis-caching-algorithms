package main

import (
	"context"
	"fmt"
	"log"

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

	res, err := client.Get(ctx, "fifocache:key").Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}
