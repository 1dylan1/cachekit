package cachekit

import (
	"hash/fnv"
	"time"
)

type ShardedCache struct {
	shards     []*Cache
	shardCount int
}

func NewShardedCache(shardCount int, expirationTime time.Duration, cleanupTime time.Duration) *ShardedCache {
	shards := make([]*Cache, shardCount)
	for i := 0; i < shardCount; i++ {
		shards[i] = New(expirationTime, cleanupTime)
	}
	return &ShardedCache{
		shards:     shards,
		shardCount: shardCount,
	}
}

func (shardedCache *ShardedCache) getShard(key string) *Cache {
	hash := fnv.New32()
	hash.Write([]byte(key))
	shardIndex := hash.Sum32() % uint32(shardedCache.shardCount)
	return shardedCache.shards[shardIndex]
}

func (shardedCache *ShardedCache) Get(key string) (interface{}, bool) {
	shard := shardedCache.getShard(key)
	return shard.Get(key)
}

func (shardedCache *ShardedCache) GetWithExpiration(key string) (interface{}, time.Time, bool) {
	shard := shardedCache.getShard(key)
	return shard.GetWithExpiration(key)
}

func (shardedCache *ShardedCache) Add(key string, content interface{}, expirationTime time.Duration) error {
	shard := shardedCache.getShard(key)
	return shard.Add(key, content, expirationTime)
}

func (shardedCache *ShardedCache) Set(key string, content interface{}, expirationTime time.Duration) {
	shard := shardedCache.getShard(key)
	shard.Set(key, content, expirationTime)
}

func (shardedCache *ShardedCache) Delete(key string) error {
	shard := shardedCache.getShard(key)
	return shard.Delete(key)
}

func (shardedCache *ShardedCache) Flush() {
	for i := 0; i < shardedCache.shardCount; i++ {
		shardedCache.shards[i].Flush()
	}
}
