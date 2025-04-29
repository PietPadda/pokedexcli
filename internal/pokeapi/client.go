// internal/pokeapi/client.go
// for the PokeAPI location areas endpoint
package pokeapi // our internal package pokeapi

import (
	// standard Go libraries
	"encoding/json" // for unmarshalling json to Go readable
	"fmt"           // for Errorf printing
	"io"            // for reading raw json data
	"net/http"      // for HTTP requests/responses

	// internal packages
	"github.com/PietPadda/pokedexcli/internal/pokecache" // our internal package pokecache
)

// API ENDPOINT STRUCTS

// LOCATION STRUCTS
// pokeapi json response struct (LAR) -- all fields exportable
type LocationAreaResponse struct {
	Count    int            `json:"count"`    // no of locations
	Next     *string        `json:"next"`     // ptr because can be null
	Previous *string        `json:"previous"` // ptr because can be null
	Results  []LocationArea `json:"results"`  // name and url array inside response (LA)
}

// single location area from the LAR struct (LA) -- all fields exportable
type LocationArea struct {
	Name string `json:"name"` // location name
	URL  string `json:"url"`  // location api url
}

// LOCATION DETAILS STRUCTS
// pokeapi json response struct (LAD) -- all fields exportable
type LocationAreaDetails struct {
	Name              string             `json:"name"`               // location name
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"` // ARRAY of pokemons found at location
}

// pokemon encounter array (PE) -- all fields exportable
type PokemonEncounter struct {
	Pokemon Pokemon `json:"pokemon"` // ARRAY of pokemon details
	// note: struct literal of a SINGLE pokemon
}

// pokemon details array from (PK) -- all fields exportable
type Pokemon struct {
	Name string `json:"name"` // pokemon name
	URL  string `json:"url"`  // pokemon api url
}

// CLIENT STRUCTS:
// Client is the PokeAPI client
type Client struct {
	httpClient http.Client      // holds HTTP client to make API requests
	cache      *pokecache.Cache // cached entries to prevent unnecessary API requests
}

// NewClient creates a new PokeAPI client
// now takes the cache for checking cached items
func NewClient(cache *pokecache.Cache) Client { // init and returns a client
	return Client{
		httpClient: http.Client{}, // init with a default HTTP client
		cache:      cache,         // init with the cache
	}
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
	res, err := c.httpClient.Do(req)

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
	res, err := c.httpClient.Do(req)

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
