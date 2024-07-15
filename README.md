# Cachekit

Cachekit is a lightweight, thread-safe, in-memory key/value store and caching solution for Go applications. It provides a simple and efficient way to store and retrieve objects with optional expiration times. The package now includes a sharded cache system for improved performance and scalability.

## Features

- Thread-safe operations for concurrent access
- Flexible expiration times for cached items
- Support for storing any Go object (interface{})
- Optional default expiration time for the entire cache
- Easy-to-use API for adding, retrieving, and removing cache entries
- Sharded cache system for better performance with high concurrency

## Use Cases

Cachekit is ideal for:
- Temporary data storage in web applications
- Caching expensive database queries or API calls
- Implementing rate limiting or throttling mechanisms
- Storing session data
- Any scenario requiring fast, in-memory data access with automatic expiration
- High-concurrency environments where a sharded cache can improve performance

## Basic Usage

```go
import (
    "github.com/1dylan1/cachekit"
    "time"
)

// Create a new cache with a default expiration time of 5 minutes for an entry, and a routine cleanup time of 10 minutes
cache := cachekit.New(5 * time.Minute, 10 * time.Minute)

// Add an item to the cache with the default expiration time we initially set
cache.Add("key", "value", cache.DefaultExpirationTime)

//Update existing value or add new value
cache.Set("key", "value",  cache.NoExpirationTime)

// Retrieve an item from the cache
value, found := cache.Get("key")
if found {
    // Use the value
}

// Remove an item from the cache
cache.Delete("key")

// 'Flush' the cache, removing all entries from it, becoming empty
cache.Flush()
```

## Sharded Cache Usage

The sharded cache system provides better performance for high-concurrency scenarios
by distributing cahce entries across multiple shards. The sharded cache provides the same API as the regular, non-sharded single cache.

```go
import (
    "github.com/1dylan1/cachekit"
    "time"
)

// Create a new sharded cache with 16 shards, default expiration time of 5 minutes, and cleanup time of 10 minutes
shardedCache := cachekit.NewShardedCache(16, 5*time.Minute, 10*time.Minute)

// Add an item to the sharded cache
shardedCache.Add("key", "value", 5*time.Minute)

// Retrieve an item from the sharded cache
value, found := shardedCache.Get("key")
if found {
    // Use the value
}

// Remove an item from the sharded cache
shardedCache.Delete("key")

// Flush all shards in the cache
shardedCache.Flush()
```