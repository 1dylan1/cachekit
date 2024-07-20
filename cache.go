package cachekit

import (
	"fmt"
	"slices"
	"sync"
	"time"
)

// Util function for the eviction strategy, removes a value by string from the slice and returns the new modified slice.
func removeByValue(s []string, value string) []string {
	index := slices.Index(s, value)
	if index != -1 {
		return append(s[:index], s[index+1:]...)
	}
	return s
}

const (
	DefaultExpirationTime time.Duration = 0
	NoExpirationTime      time.Duration = -1
)

type Entry struct {
	Content             interface{}
	ExpirationTimestamp int64
}

type cache struct {
	expirationTime  time.Duration
	cleanupTime     time.Duration
	nextCleanupTime int64
	expiringKeys    []string
	entries         map[string]Entry
	mutex           sync.RWMutex
}

func New(expirationTime time.Duration, cleanupTime time.Duration) *cache {
	entries := make(map[string]Entry)
	cache := &cache{
		expirationTime:  expirationTime,
		cleanupTime:     cleanupTime,
		nextCleanupTime: time.Now().Add(cleanupTime).Unix(),
		expiringKeys:    []string{},
		entries:         entries,
		mutex:           sync.RWMutex{},
	}

	if cleanupTime != 0 && cleanupTime != NoExpirationTime {
		go cache.cleanUpCache()
	}

	return cache
}

func (cache *cache) cleanUpCache() {
	ticket := time.NewTicker(cache.cleanupTime)
	for range ticket.C {
		cache.removeExpiredEntries()
		cache.nextCleanupTime = time.Now().Add(cache.cleanupTime).Unix()
	}
}

func (cache *cache) removeExpiredEntries() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	currentTimestamp := time.Now().Unix()
	var remainingKeys []string

	for _, key := range cache.expiringKeys {
		entry, exists := cache.entries[key]
		if !exists {
			continue
		}

		if entry.ExpirationTimestamp < currentTimestamp && entry.ExpirationTimestamp != int64(NoExpirationTime) {
			delete(cache.entries, key)
		} else {
			remainingKeys = append(remainingKeys, key)
		}
	}

	cache.expiringKeys = remainingKeys
}

/*
Returns the entry from the cache from given key, and a bool if it was found or not.
If the key has no object associated with it, or wasn't found, returns nil.
*/
func (cache *cache) Get(key string) (interface{}, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	entry, found := cache.entries[key]

	if !found {
		return nil, false
	}

	return entry.Content, true
}

/*
Returns the entry from the cache with given key, the time of the expiration, and bool if it was found or not.
If the key has no object associated with it, or wasn't found, returns nil.
*/
func (cache *cache) GetWithExpiration(key string) (interface{}, time.Time, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	entry, found := cache.entries[key]

	if !found {
		return nil, time.Time{}, false
	}

	return entry.Content, time.Unix(entry.ExpirationTimestamp, 0), true
}

/*
Tries to add a key/entry into the cache.
If the cache already has that key it will throw an error. If you'd like to add (or "update") a value even if its key exists, use the Set method
*/
func (cache *cache) Add(key string, content interface{}, expirationTime time.Duration) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	_, found := cache.entries[key]
	if found {
		return fmt.Errorf("[cachekit] Entry %s already exists in cache", key)
	}
	expiration := time.Now().Add(expirationTime).Unix()

	if expirationTime == DefaultExpirationTime {
		expiration = time.Now().Add(cache.expirationTime).Unix()
	}

	if expiration <= cache.nextCleanupTime {
		cache.expiringKeys = append(cache.expiringKeys, key)
	}

	entry := Entry{
		Content:             content,
		ExpirationTimestamp: expiration,
	}

	cache.entries[key] = entry
	return nil
}

/*
Adds or updates a key/entry to the cache, ignoring if the key already exists or not.
*/
func (cache *cache) Set(key string, content interface{}, expirationTime time.Duration) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	expiration := time.Now().Add(expirationTime).Unix()

	if expirationTime == DefaultExpirationTime {
		expiration = time.Now().Add(cache.expirationTime).Unix()
	}

	if slices.Contains(cache.expiringKeys, key) {
		cache.expiringKeys = removeByValue(cache.expiringKeys, key)
	}

	if expiration <= cache.nextCleanupTime {
		cache.expiringKeys = append(cache.expiringKeys, key)
	}

	entry := Entry{
		Content:             content,
		ExpirationTimestamp: expiration,
	}

	cache.entries[key] = entry
}

func (cache *cache) Delete(key string) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	_, found := cache.entries[key]

	if !found {
		return fmt.Errorf("[cachekit] could not find key %s in cache", key)
	}

	if slices.Contains(cache.expiringKeys, key) {
		cache.expiringKeys = removeByValue(cache.expiringKeys, key)
	}

	delete(cache.entries, key)

	return nil
}

/*
Cleans the cache of all entries
*/
func (cache *cache) Flush() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.entries = make(map[string]Entry)
	cache.expiringKeys = []string{}
}

/*
Returns the number of entries in the cache.
*/
func (cache *cache) Length() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return len(cache.entries)
}
