// pokecache_test.go
package pokecache

import (
	"fmt"
	"testing"
	"time"
)

// importing testing package for unit tests

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.CacheAdd(c.key, c.val)
			val, ok, err := cache.CacheGet(c.key)
			if err != nil {
				t.Errorf("CacheAdd unsuccesful")
				return
			}
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected to find value")
				return
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.CacheAdd("https://example.com", []byte("testdata"))

	_, ok, err := cache.CacheGet("https://example.com")
	if err != nil {
		t.Errorf("CacheAdd unsuccesful")
		return
	}
	if !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime)

	_, ok, err = cache.CacheGet("https://example.com")
	if err != nil {
		t.Errorf("CacheAdd unsuccesful")
		return
	}
	if ok {
		t.Errorf("expected to not find key")
		return
	}
}
