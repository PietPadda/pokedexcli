// internal/pokecache/pokecache.go
// for caching and reaping PokeAPI responses
package pokecache

import (
	"fmt"
	"sync" // for mutex concurrency (maps aren't thread safe)
	"time" // required for Timer functions
)

// cache entries struct for time created and raw data
// lowercase (private) as its internal use only
type cacheEntry struct {
	createdAt time.Time // time at which cache entry was created
	val       []byte    // raw data storage
}

// cache entries map, mutex for map concurrency and reaper duration
// capped (public) for exposing to other packages
type Cache struct {
	cache    map[string]cacheEntry // map of cache entries
	mu       *sync.Mutex           // mutex since maps aren't thread safe (must init in constructor as its ptr)
	interval time.Duration         // store the duration here, which NewCache accepts as param and stores here
}

// constructor function for making new cache
// takes interval as arg, inits new cache, starts reaper goroutine and returns the cache
// capped (public) for exposing to other packages
func NewCache(interval time.Duration) *Cache { // ptr = more efficient, no data copying when passing
	cache := &Cache{
		cache:    make(map[string]cacheEntry), // inits new cache
		mu:       &sync.Mutex{},               // inits the mutex (safe, avoid nil ptr deref)
		interval: interval,                    // takes the interval and stores in cache return
	}
	go cache.reapLoop() // starts "reaper" goroutine
	return cache        // return the cache
}

// reapLoop method to remove old cache entries for memory efficiency
// sep goroutine, locks mutex, checks ages of entries and cleans, then unlocks mutex
// runs periodically via time.Ticker interval
func (c *Cache) reapLoop() {
	// init the time ticker and goroutine
	ticker := time.NewTicker(c.interval) // init new ticker with age limit duration
	defer ticker.Stop()                  // makes ticker stop on function loop ending

	// infinite reapLoop based on timer
	for {
		// wait for the next tick
		// this blocks until the ticker sends the next value on the channel
		<-ticker.C // time.Ticker has its own channel called C
		// <- = reads from the ticker channel, which sends the current time at regular intervals
		// we set the "tick" to be c.interval, which is the cache age limit

		// lock mutex before accessing map
		c.mu.Lock()
		// defer c.mu.Unlock() - dont defer as its infinite loop!

		// loop thru cacheEntry keys and their values
		for k, v := range c.cache {
			ageCache := v.createdAt // get key-value age of cache
			ageLimit := c.interval  // set simpel var for readability

			// reap check based on cache age
			if time.Since(ageCache) >= ageLimit {
				delete(c.cache, k) // delete the cache entry by it's createdAt time
			}
		}

		// unlock mutex after accessing map
		c.mu.Unlock()
	}
}

// cache add function -- adds a new entry to the cache
// takes *Cache -- update the actual cache map NOT a copy
// takes a URL-key:DATA-value pair as input
func (c *Cache) CacheAdd(key string, val []byte) error { // returns new cache
	// nil ptr check
	if c == nil {
		return fmt.Errorf("cacheAdd called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// get inputs (just for readability)
	url := key  // location url as map key (identifies the entry)
	data := val // location data as map key value

	// lock mutex before accessing map
	c.mu.Lock()
	defer c.mu.Unlock() // will unlock on *Cache return

	// create new cache entry
	entry := cacheEntry{
		createdAt: time.Now(), // time.Now() = get current time for createdAt
		val:       data,
	}

	// update existing cacheEntry map
	c.cache[url] = entry // fetches the whole struct and updates timestamp and data

	// successfully  added new cache entry
	return nil
}

// cache get function -- gets an existing entry from the cache
// takes *Cache -- returns a []byte and "found" bool
// takes a URL-key as input
func (c *Cache) CacheGet(key string) ([]byte, bool, error) { // returns existing cache
	// nil ptr check
	if c == nil {
		return nil, false, fmt.Errorf("cacheGet called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// get inputs (just for readability)
	url := key // location url as map key (identifies the entry)

	// lock mutex before accessing map
	c.mu.Lock()
	defer c.mu.Unlock() // will unlock on *Cache return

	// loop thru cache to see if url can be found
	entry, ok := c.cache[url]

	// exist check
	if !ok {
		return nil, false, nil // not found, no error
	}

	// otherwise, found entry and return as success
	data := entry.val // get entry's val field, []byte
	return data, true, nil
}
