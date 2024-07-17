package cachekit

import (
	"testing"
	"time"

	"github.com/1dylan1/cachekit"
)

func TestNewCache(t *testing.T) {
	testCache := cachekit.New(cachekit.DefaultExpirationTime, cachekit.NoExpirationTime)

	nonExistentA, found := testCache.Get("foo")
	if found || nonExistentA != nil {
		t.Error("Getting value from cache that shouldn't exist: ", nonExistentA)
	}

	nonExistentB, found := testCache.Get("bar")
	if found || nonExistentB != nil {
		t.Error("Getting value from cache that shouldn't exist: ", nonExistentB)
	}

	testCache.Add("john", "doe", cachekit.NoExpirationTime)

	entry, found := testCache.Get("john")
	if !found || entry == nil {
		t.Error("Didnt find key in cache that should exist!")
	}

	testCache.Add("jane", "doe", cachekit.NoExpirationTime)

	entryB, found := testCache.Get("jane")
	if !found || entryB == nil {
		t.Error("Didnt find key in cache that should exist!")
	}

	testCache.Set("jack", 10.51, cachekit.NoExpirationTime)

	entryC, found := testCache.Get("jack")
	if !found || entryC == nil {
		t.Error("Didnt find key in cache that should exist!")
	}

	testCache.Delete("jane")
	entryDelete, found := testCache.Get("jane")
	if found || entryDelete != nil {
		t.Error("Found entry that shouldnt exist")
	}
}

func TestCacheTimeDurations(t *testing.T) {

	testCache := cachekit.New(time.Millisecond*1, time.Millisecond*5)

	testCache.Set("a", 0, cachekit.DefaultExpirationTime)
	testCache.Set("b", 1, cachekit.NoExpirationTime)
	testCache.Set("c", 2, time.Millisecond*2)

	<-time.After(time.Second * 1)
	_, found := testCache.Get("a")
	if found {
		t.Error("Found a when it should have been cleaned up")
	}
}

func TestCacheFlush(t *testing.T) {

	testCache := cachekit.New(cachekit.DefaultExpirationTime, cachekit.NoExpirationTime)

	testCache.Set("a", 1, cachekit.DefaultExpirationTime)
	testCache.Set("b", 1, cachekit.DefaultExpirationTime)
	testCache.Set("c", 1, cachekit.DefaultExpirationTime)
	testCache.Flush()
	entries := testCache.Length()
	if entries != 0 {
		t.Error("Cache flush did not remove all items, length:", entries)
	}
}
