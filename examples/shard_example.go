package examples

import (
	"cachekit"
	"fmt"
	"time"
)

func ShardCacheExample() {

	// Create a new sharded cache with 16 shards, default expiration time of 5 minutes, and cleanup time of 10 minutes
	shardedCache := cachekit.NewShardedCache(16, 5*time.Minute, 10*time.Minute)

	// Add an item to the sharded cache
	shardedCache.Add("key", "value", 5*time.Minute)

	// Retrieve an item from the sharded cache
	value, found := shardedCache.Get("key")
	if found {
		fmt.Println(value)
	}

	// Remove an item from the sharded cache
	shardedCache.Delete("key")

	// Flush all shards in the cache
	shardedCache.Flush()
}
