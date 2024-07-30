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

type entry struct {
	Content             interface{}
	ExpirationTimestamp int64
}

type Cache struct {
	expirationTime  time.Duration
	cleanupTime     time.Duration
	nextCleanupTime int64
	expiringKeys    []string
	entries         map[string]entry
	mutex           sync.RWMutex
}

func New(expirationTime time.Duration, cleanupTime time.Duration) *Cache {
	entries := make(map[string]entry)
	Cache := &Cache{
		expirationTime:  expirationTime,
		cleanupTime:     cleanupTime,
		nextCleanupTime: time.Now().Add(cleanupTime).Unix(),
		expiringKeys:    []string{},
		entries:         entries,
		mutex:           sync.RWMutex{},
	}

	if cleanupTime != 0 && cleanupTime != NoExpirationTime {
		go Cache.cleanUpCache()
	}

	return Cache
}

func (Cache *Cache) cleanUpCache() {
	ticket := time.NewTicker(Cache.cleanupTime)
	for range ticket.C {
		Cache.removeExpiredEntries()
	}
}

func (Cache *Cache) removeExpiredEntries() {
	Cache.mutex.Lock()
	defer Cache.mutex.Unlock()

	currentTimestamp := time.Now().Unix()
	var remainingKeys []string

	for _, key := range Cache.expiringKeys {
		entry, exists := Cache.entries[key]
		if !exists {
			continue
		}

		if entry.ExpirationTimestamp < currentTimestamp && entry.ExpirationTimestamp != int64(NoExpirationTime) {
			delete(Cache.entries, key)
		} else {
			remainingKeys = append(remainingKeys, key)
		}
	}

	Cache.expiringKeys = remainingKeys
}

/*
Returns the entry from the Cache from given key, and a bool if it was found or not.
If the key has no object associated with it, or wasn't found, returns nil.
*/
func (Cache *Cache) Get(key string) (interface{}, bool) {
	Cache.mutex.RLock()
	defer Cache.mutex.RUnlock()

	entry, found := Cache.entries[key]

	if !found {
		return nil, false
	}

	return entry.Content, true
}

/*
Returns the entry from the Cache with given key, the time of the expiration, and bool if it was found or not.
If the key has no object associated with it, or wasn't found, returns nil.
*/
func (Cache *Cache) GetWithExpiration(key string) (interface{}, time.Time, bool) {
	Cache.mutex.RLock()
	defer Cache.mutex.RUnlock()

	entry, found := Cache.entries[key]

	if !found {
		return nil, time.Time{}, false
	}

	return entry.Content, time.Unix(entry.ExpirationTimestamp, 0), true
}

/*
Tries to add a key/entry into the Cache.
If the Cache already has that key it will throw an error. If you'd like to add (or "update") a value even if its key exists, use the Set method
*/
func (Cache *Cache) Add(key string, content interface{}, expirationTime time.Duration) error {
	Cache.mutex.Lock()
	defer Cache.mutex.Unlock()
	_, found := Cache.entries[key]
	if found {
		return fmt.Errorf("[Cachekit] entry %s already exists in Cache", key)
	}
	expiration := time.Now().Add(expirationTime).Unix()

	if expirationTime == DefaultExpirationTime {
		expiration = time.Now().Add(Cache.expirationTime).Unix()
	}

	if expiration <= Cache.nextCleanupTime {
		Cache.expiringKeys = append(Cache.expiringKeys, key)
	}

	entry := entry{
		Content:             content,
		ExpirationTimestamp: expiration,
	}

	Cache.entries[key] = entry
	return nil
}

/*
Adds or updates a key/entry to the Cache, ignoring if the key already exists or not.
*/
func (Cache *Cache) Set(key string, content interface{}, expirationTime time.Duration) {
	Cache.mutex.Lock()
	defer Cache.mutex.Unlock()

	expiration := time.Now().Add(expirationTime).Unix()

	if expirationTime == DefaultExpirationTime {
		expiration = time.Now().Add(Cache.expirationTime).Unix()
	}

	if slices.Contains(Cache.expiringKeys, key) {
		Cache.expiringKeys = removeByValue(Cache.expiringKeys, key)
	}

	if expiration <= Cache.nextCleanupTime {
		Cache.expiringKeys = append(Cache.expiringKeys, key)
	}

	entry := entry{
		Content:             content,
		ExpirationTimestamp: expiration,
	}

	Cache.entries[key] = entry
}

func (Cache *Cache) Delete(key string) error {
	Cache.mutex.Lock()
	defer Cache.mutex.Unlock()

	_, found := Cache.entries[key]

	if !found {
		return fmt.Errorf("[Cachekit] could not find key %s in Cache", key)
	}

	if slices.Contains(Cache.expiringKeys, key) {
		Cache.expiringKeys = removeByValue(Cache.expiringKeys, key)
	}

	delete(Cache.entries, key)

	return nil
}

/*
Cleans the Cache of all entries
*/
func (Cache *Cache) Flush() {
	Cache.mutex.Lock()
	defer Cache.mutex.Unlock()

	Cache.entries = make(map[string]entry)
	Cache.expiringKeys = []string{}
}

/*
Returns the number of entries in the Cache.
*/
func (Cache *Cache) Length() int {
	Cache.mutex.RLock()
	defer Cache.mutex.RUnlock()
	return len(Cache.entries)
}
