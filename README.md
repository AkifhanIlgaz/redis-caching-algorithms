# redis-caching-algorithms

This project provides Go implementations of several common cache eviction algorithms. These algorithms are fundamental for managing memory in caching systems, such as Redis, by deciding which items to discard when the cache is full.

## Algorithms

- FIFO (First-In, First-Out)
- LFU (Least Frequently Used)
- LRU (Least Recently Used)
- TTL (Time-To-Live)

## Usage

## Implementation Details

All caching algorithms are implemented using a `Cache` interface that defines the following methods:

- `Get(id string) (User, error)`: Retrieves a user from the cache.
- `Set(user User) error`: Adds a user to the cache.
- `Delete(key string) error`: Removes a key from the cache.
- `CacheSize() int`: Returns the current number of items in the cache.

The `User` struct is defined as follows:

```go
type User struct {
    Id   string `json:"id" redis:"-"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}
```

### FIFO (First-In, First-Out)

The FIFO cache is implemented using a Redis list to maintain the order of items. When the cache is full, the oldest item is removed from the left of the list.

### LFU (Least Frequently Used)

The LFU cache is implemented using a Redis sorted set to track the frequency of access. The score of each member in the sorted set represents the frequency of access. When an item is accessed, its score is incremented. When the cache is full, the item with the lowest score is removed.

### LRU (Least Recently Used)

The LRU cache is implemented using a Redis sorted set to maintain the order of items by their last access time. The score of each member in the sorted set represents the timestamp of the last access. When an item is accessed, its score is updated to the current time. When the cache is full, the item with the lowest score (oldest timestamp) is removed.

### TTL (Time-To-Live)

The TTL cache is implemented using Redis's built-in key expiration feature. When a new item is added to the cache, it is set with a specific time-to-live (TTL). Redis automatically removes the item from the cache when its TTL has expired. This approach is ideal for data that becomes stale or irrelevant after a certain period.

## Usage

To see the caching algorithms in action, you can run the `test.go` file in the `cmd/test` directory. This will demonstrate the step-by-step execution of the cache logic.

## Todos

- [ ] **Add More Caching Algorithms**: Implement other caching strategies like MRU (Most Recently Used) or RR (Random Replacement).
- [ ] **Unit Tests**: Develop a comprehensive test suite to verify the correctness of each caching algorithm.
- [ ] **Generic Cache Interface**: Refactor the `Cache` interface to be more generic, allowing it to store different data types, not just `User` structs.
- [ ] **Configuration**: Allow cache parameters (like size, TTL) to be configured through a file or environment variables.
- [ ] **Improved Example**: Enhance the example in `cmd/test` to be more interactive or to simulate a more realistic use case.
