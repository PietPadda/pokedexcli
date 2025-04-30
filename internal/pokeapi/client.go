// internal/pokeapi/client.go
// for the PokeAPI location areas endpoint
package pokeapi // our internal package pokeapi

import (
	// standard Go libraries
	"encoding/json" // for unmarshalling json to Go readable
	"fmt"           // for Errorf printing
	"io"            // for reading raw json data
	"net/http"      // for HTTP requests/responses
	"sync"          // for Mutex on map concurrency safety

	// internal packages
	"github.com/PietPadda/pokedexcli/internal/pokecache" // our internal package pokecache
)

// API ENDPOINT STRUCTS

// LOCATION STRUCTS
// pokeapi json response struct (LAR) -- all fields exportable
type LocationAreaResponse struct {
	Results  []LocationArea `json:"results"`  // name and url array inside response (LA)
	Next     *string        `json:"next"`     // ptr because can be null
	Previous *string        `json:"previous"` // ptr because can be null
	Count    int            `json:"count"`    // no of locations
}

// single location area from the LAR struct (LA) -- all fields exportable
type LocationArea struct {
	Name string `json:"name"` // location name
	URL  string `json:"url"`  // location api url
}

// LOCATION DETAILS STRUCTS
// pokeapi json response struct (LAD) -- all fields exportable
type LocationAreaDetails struct {
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"` // ARRAY of pokemons found at location
	Name              string             `json:"name"`               // location name
}

// pokemon encounter array (PE) -- all fields exportable
type PokemonEncounter struct {
	Pokemon Pokemon `json:"pokemon"` // struct literal of SINGLE pokemon detail
}

// pokemon details array (PK) -- all fields exportable
type Pokemon struct {
	Name string `json:"name"` // pokemon name
	URL  string `json:"url"`  // pokemon api url
}

// POKEMON STATS STRUCTS
// pokemon stats (PS) -- all fields exportable
type PokemonStats struct {
	Name           string `json:"name"`            // pokemon name (for storing in pokedex)
	BaseExperience int    `json:"base_experience"` // pokemon base experience (for catch probability)
	ID             int    `json:"id"`              // pokemon id (we use name, but can also use id)
}

// CLIENT STRUCTS:
// Client is the PokeAPI client
type Client struct {
	PokeapiClient http.Client      // holds HTTP client to make API requests
	cache         *pokecache.Cache // cached entries to prevent unnecessary API requests
}

// NewClient creates a new PokeAPI client
// now takes the cache for checking cached items
func NewClient(cache *pokecache.Cache) Client { // init and returns a client
	return Client{
		PokeapiClient: http.Client{}, // init with a default HTTP client
		cache:         cache,         // init with the cache
	}
}

// POKEDEX STRUCTS:
// Pokedex is where the store and inspect the pokemon we catch
// capped (public) for exposing to other packages
type Pokedex struct {
	pokemon map[string]PokemonStats // map of pokedex entries
	mu      *sync.RWMutex           // mutex since maps aren't thread safe (must init in constructor as its ptr)
}

// CORE: we use RWMutex here as we will frequently be reading from but, it STILL allows exclusive writing
// performance boost for repeatedly using it!

// constructor function for making new Pokedex
// inits new Pokedex, and returns the Pokedex
// capped (public) for exposing to other packages
func NewPokedex() *Pokedex { // ptr = more efficient, no data copying when passing
	pokedex := &Pokedex{
		pokemon: make(map[string]PokemonStats), // inits new pokedex
		mu:      &sync.RWMutex{},               // inits the mutex (safe, avoid nil ptr deref)
	}
	return pokedex // return the pokedex
}

// function to get locations for the PokeAPI client
// takes a url request input, and outputs the location area and success/failure error
// it's a method on the client (Go style "OOP")
func (c *Client) GetLocationAreas(pageURL string) (LocationAreaResponse, error) {
	// nil ptr check
	if c == nil {
		return LocationAreaResponse{}, fmt.Errorf("GetLocationAreas called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// determine default url for locations
	baseURL := "https://pokeapi.co/api/v2" // api url
	resourceURL := "/location-area"        // resource url
	fullURL := baseURL + resourceURL       // full url

	// handle empty input url
	if pageURL == "" {
		pageURL = fullURL // set url to fullURL
	}

	// cached entry call, store IF found and IF error
	cachedEntries, ok, err := c.cache.CacheGet(pageURL) // if response already cached

	// cache entries call check
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("error getting cached entries: %w", err) // nil slice & error
	}

	// if cache entries found
	if ok {
		// first, create nil slice for external data response
		var locationRes LocationAreaResponse

		// then unmarshal
		err := json.Unmarshal(cachedEntries, &locationRes)

		// unmarshal to conv from raw json to go readable code
		if err != nil {
			return locationRes, fmt.Errorf("error unmarshalling json data: %w", err) // nil slice & error
		}

		// can now return the CACHED location area response from server as success
		return locationRes, nil // nil error
	}

	// if not cached, need to make new HTTP GET request

	// HTTP GET request using newrequest for more flexibility
	req, err := http.NewRequest("GET", pageURL, nil) // GET request, so no response body

	// HTTP request check
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("error with HTTP request: %w", err) // empty slice & error
	}

	// modify GET request header (not required, but BEST GO PRACTICE)
	req.Header.Set("Accept", "application/json") // expects json data as HTTP response
	// CORE: "Content-Type" - sending TO server, "Accept" - response FROM server

	// don’t create a new HTTP client
	// client := &http.Client{}

	// we no longer create a new client, but use the pokeapi httpClient below
	// client do GET request
	res, err := c.PokeapiClient.Do(req)

	// client do GET check
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("error client doing request: %w", err) // empty slice & error
	}

	// defer to close network connectoin after reading to prevent mem leak
	defer res.Body.Close()

	// get server response status code
	statusCode := res.StatusCode // server response status code
	resStatus := res.Status      // status code AND description

	// status code check
	if statusCode != http.StatusOK { // if not 200
		return LocationAreaResponse{}, fmt.Errorf("error server response status code unsuccesful: %s", resStatus) // empty slice & status code w descr
	}

	// read server response body as raw json data,[]byte slice
	body, err := io.ReadAll(res.Body)

	// read body check
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("error reading server response body: %w", err) // empty slice & error
	}

	// create nil slice for external data response
	var locationRes LocationAreaResponse

	// unmarshal to conv from raw json to go readable code
	err = json.Unmarshal(body, &locationRes)

	// unmarshal check
	if err != nil {
		return locationRes, fmt.Errorf("error unmarshalling json data: %w", err) // nil slice & error
	}

	// the http response is now unmarshalled, let's first add it to the cache for future reference!
	err = c.cache.CacheAdd(pageURL, body) // add url as key to cache + body (the raw "data"), return error

	// cache add check
	if err != nil {
		fmt.Printf("error adding to cache: %v\n", err) // HTTP request slice & error
		// DON'T RETURN! we still want to continue with the actual HTTP response return, else nothing happens!
	} // printf for FORMATTED print, println can't use %v verb!

	// can now return the location area response from server as success
	return locationRes, nil // nil error
}

// function to get details of a location using the PokeAPI client
// takes a location name request input, and outputs the location area details and success/failure error
// it's a method on the client (Go style "OOP")
func (c *Client) GetLocationArea(locationName string) (LocationAreaDetails, error) {
	// nil ptr check
	if c == nil {
		return LocationAreaDetails{}, fmt.Errorf("GetLocationArea called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// locationname check
	if locationName == "" {
		return LocationAreaDetails{}, fmt.Errorf("location name cannot be empty") // early return
	}

	// determine default url for locations
	baseURL := "https://pokeapi.co/api/v2"         // api url
	endpointURL := "/location-area/"               // api endpoint url
	resourceURL := locationName                    // location name
	fullURL := baseURL + endpointURL + resourceURL // full url

	// cached entry call, store IF found and IF error
	cachedEntries, ok, err := c.cache.CacheGet(fullURL) // if response already cached

	// cache entries call check
	if err != nil {
		return LocationAreaDetails{}, fmt.Errorf("error getting cached entries: %w", err) // nil slice & error
	}

	// if cache entries found
	if ok {
		// first, create nil slice for external data response
		var locationRes LocationAreaDetails

		// then unmarshal
		err := json.Unmarshal(cachedEntries, &locationRes)

		// unmarshal to conv from raw json to go readable code
		if err != nil {
			return locationRes, fmt.Errorf("error unmarshalling json data: %w", err) // nil slice & error
		}

		// can now return the CACHED location area response from server as success
		return locationRes, nil // nil error
	}

	// if not cached, need to make new HTTP GET request

	// HTTP GET request using newrequest for more flexibility
	req, err := http.NewRequest("GET", fullURL, nil) // GET request, so no response body

	// HTTP request check
	if err != nil {
		return LocationAreaDetails{}, fmt.Errorf("error with HTTP request: %w", err) // empty slice & error
	}

	// modify GET request header (not required, but BEST GO PRACTICE)
	req.Header.Set("Accept", "application/json") // expects json data as HTTP response
	// CORE: "Content-Type" - sending TO server, "Accept" - response FROM server

	// client do GET request using pokeapi client
	res, err := c.PokeapiClient.Do(req)

	// client do GET check
	if err != nil {
		return LocationAreaDetails{}, fmt.Errorf("error client doing request: %w", err) // empty slice & error
	}

	// defer to close network connectoin after reading to prevent mem leak
	defer res.Body.Close()

	// get server response status code
	statusCode := res.StatusCode // server response status code
	resStatus := res.Status      // status code AND description

	// status code check
	if statusCode != http.StatusOK { // if not 200
		return LocationAreaDetails{}, fmt.Errorf("error server response status code unsuccesful: %s", resStatus) // empty slice & status code w descr
	}

	// read server response body as raw json data,[]byte slice
	body, err := io.ReadAll(res.Body)

	// read body check
	if err != nil {
		return LocationAreaDetails{}, fmt.Errorf("error reading server response body: %w", err) // empty slice & error
	}

	// create nil slice for external data response
	var locationRes LocationAreaDetails

	// unmarshal to conv from raw json to go readable code
	err = json.Unmarshal(body, &locationRes)

	// unmarshal check
	if err != nil {
		return locationRes, fmt.Errorf("error unmarshalling json data: %w", err) // nil slice & error
	}

	// the http response is now unmarshalled, let's first add it to the cache for future reference!
	err = c.cache.CacheAdd(fullURL, body) // add url as key to cache + body (the raw "data"), return error

	// cache add check
	if err != nil {
		fmt.Printf("error adding to cache: %v\n", err) // HTTP request slice & error
		// DON'T RETURN! we still want to continue with the actual HTTP response return, else nothing happens!
	} // printf for FORMATTED print, println can't use %v verb!

	// can now return the location area DETAILS response from server as success
	return locationRes, nil // nil error
}

// function to get stats of a pokemon using the PokeAPI client
// takes a pokemon name request input, and outputs the pokemon stats and success/failure error
// it's a method on the client (Go style "OOP")
func (c *Client) GetPokemonStats(pokemonName string) (PokemonStats, error) {
	// nil ptr check
	if c == nil {
		return PokemonStats{}, fmt.Errorf("GetPokemonStats called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// pokemon name check
	if pokemonName == "" {
		return PokemonStats{}, fmt.Errorf("pokemon name cannot be empty") // early return
	}

	// determine default url for locations
	baseURL := "https://pokeapi.co/api/v2"         // api url
	endpointURL := "/pokemon/"                     // api endpoint url
	resourceURL := pokemonName                     // pokemon name
	fullURL := baseURL + endpointURL + resourceURL // full url
	// reference: GET https://pokeapi.co/api/v2/pokemon/{id or name}/

	// cached entry call, store IF found and IF error
	cachedEntries, ok, err := c.cache.CacheGet(fullURL) // if response already cached

	// cache entries call check
	if err != nil {
		return PokemonStats{}, fmt.Errorf("error getting cached entries: %w", err) // nil slice & error
	}

	// if cache entries found
	if ok {
		// first, create nil slice for external data response
		var pokemonRes PokemonStats

		// then unmarshal
		err := json.Unmarshal(cachedEntries, &pokemonRes)

		// unmarshal to conv from raw json to go readable code
		if err != nil {
			return pokemonRes, fmt.Errorf("error unmarshalling json data: %w", err) // nil slice & error
		}

		// can now return the CACHED location area response from server as success
		return pokemonRes, nil // nil error
	}

	// if not cached, need to make new HTTP GET request

	// HTTP GET request using newrequest for more flexibility
	req, err := http.NewRequest("GET", fullURL, nil) // GET request, so no response body

	// HTTP request check
	if err != nil {
		return PokemonStats{}, fmt.Errorf("error with HTTP request: %w", err) // empty slice & error
	}

	// modify GET request header (not required, but BEST GO PRACTICE)
	req.Header.Set("Accept", "application/json") // expects json data as HTTP response
	// CORE: "Content-Type" - sending TO server, "Accept" - response FROM server

	// client do GET request using pokeapi client
	res, err := c.PokeapiClient.Do(req)

	// client do GET check
	if err != nil {
		return PokemonStats{}, fmt.Errorf("error client doing request: %w", err) // empty slice & error
	}

	// defer to close network connectoin after reading to prevent mem leak
	defer res.Body.Close()

	// get server response status code
	statusCode := res.StatusCode // server response status code
	resStatus := res.Status      // status code AND description

	// status code check
	if statusCode != http.StatusOK { // if not 200
		return PokemonStats{}, fmt.Errorf("error server response status code unsuccesful: %s", resStatus) // empty slice & status code w descr
	}

	// read server response body as raw json data,[]byte slice
	body, err := io.ReadAll(res.Body)

	// read body check
	if err != nil {
		return PokemonStats{}, fmt.Errorf("error reading server response body: %w", err) // empty slice & error
	}

	// create nil slice for external data response
	var pokemonRes PokemonStats

	// unmarshal to conv from raw json to go readable code
	err = json.Unmarshal(body, &pokemonRes)

	// unmarshal check
	if err != nil {
		return pokemonRes, fmt.Errorf("error unmarshalling json data: %w", err) // nil slice & error
	}

	// the http response is now unmarshalled, let's first add it to the cache for future reference!
	err = c.cache.CacheAdd(fullURL, body) // add url as key to cache + body (the raw "data"), return error

	// cache add check
	if err != nil {
		fmt.Printf("error adding to cache: %v\n", err) // HTTP request slice & error
		// DON'T RETURN! we still want to continue with the actual HTTP response return, else nothing happens!
	} // printf for FORMATTED print, println can't use %v verb!

	// can now return the location area DETAILS response from server as success
	return pokemonRes, nil // nil error
}

// pokedex add function -- adds a new entry to the pokedex
// takes *Pokedex -- update the actual pokedex map NOT a copy
// takes a URL-key:DATA-value pair as input
func (p *Pokedex) PokemonAdd(name string, stats PokemonStats) error { // adds new pokemon entry
	// nil ptr check
	if p == nil {
		return fmt.Errorf("PokemonAdd called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// get inputs (just for readability)
	pokemonName := name   // pokemon name that we caught
	pokemonStats := stats // pokemon stats

	// lock mutex before accessing map
	p.mu.Lock()
	defer p.mu.Unlock() // will unlock on *Pokedex return

	// update pokedex map by adding the pokemon
	p.pokemon[pokemonName] = pokemonStats // fetches the whole struct and updates pokemon and stats
	// p is ptr to pokedex, and pokemon is the map field. We set the map key to the name and its val is the stats!

	// successfully added new pokedex entry
	return nil
}

// pokedex get function -- gets an existing entry from the pokedex
// takes *Pokedex -- returns a PokemonStats struct and "found" bool, and error
// takes a URL-key as input
func (p *Pokedex) PokemonGet(name string) (PokemonStats, bool, error) { // returns existing pokemon
	// nil ptr check
	if p == nil {
		return PokemonStats{}, false, fmt.Errorf("PokemonGet called with nil receiver") // early return
	} // runtime panic if try access ptr fields, no memory location!

	// get inputs (just for readability)
	pokemonName := name // pokemon name that we caught

	// READ lock mutex before accessing map
	p.mu.RLock()         // READ lock only, allows fast access!
	defer p.mu.RUnlock() // will READ unlock on *Pokedex return

	// loop thru pokedex map to see if pokemon can be found
	entry, ok := p.pokemon[pokemonName]

	// exist check
	if !ok {
		return PokemonStats{}, false, nil // not found, no error
	}

	// otherwise, found entry and return as success
	return entry, true, nil
}
