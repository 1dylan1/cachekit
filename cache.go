package cachekit

import (
	"fmt"
	"sync"
	"time"
)

const (
	DefaultExpirationTime time.Duration = 0
	NoExpirationTime      time.Duration = -1
)

type Entry struct {
	Content             interface{}
	ExpirationTimestamp int64
}

type cache struct {
	expirationTime time.Duration
	cleanupTime    time.Duration
	entries        map[string]Entry
	mutex          sync.RWMutex
}

func New(expirationTime time.Duration, cleanupTime time.Duration) *cache {
	entries := make(map[string]Entry)
	cache := &cache{
		expirationTime: expirationTime,
		cleanupTime:    cleanupTime,
		entries:        entries,
		mutex:          sync.RWMutex{},
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
	}
}

func (cache *cache) removeExpiredEntries() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	currentTimestamp := time.Now().Unix()
	for key, entry := range cache.entries {
		if entry.ExpirationTimestamp < currentTimestamp && entry.ExpirationTimestamp != int64(NoExpirationTime) {
			delete(cache.entries, key)
		}
	}
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
}

/*
Returns the number of entries in the cache.
*/
func (cache *cache) Length() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return len(cache.entries)
}
