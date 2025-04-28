// internal/pokeapi/client.go
// for the PokeAPI location areas endpoint
package pokeapi // our internal package pokeapi

import (
	"encoding/json" // for unmarshalling json to Go readable
	"fmt"           // for Errorf printing
	"io"            // for reading raw json data
	"net/http"      // for HTTP requests/responses
)

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

// Client is the PokeAPI client
type Client struct {
	httpClient http.Client // holds HTTP client to make API requests
}

// NewClient creates a new PokeAPI client
func NewClient() Client { // init and returns a client
	return Client{
		httpClient: http.Client{}, // init with a default HTTP client
	}
}

// function to get locations for the PokeAPI client
// takes a url request input, and outputs the location area and success/failure error
// it's a method on the client (Go style "OOP")
func (c *Client) GetLocationAreas(pageURL string) (LocationAreaResponse, error) {
	// determine default url for locations
	baseURL := "https://pokeapi.co/api/v2" // api url
	resourceURL := "/location-area"        // resource url
	fullURL := baseURL + resourceURL       // full url

	// handle empty input url
	if pageURL == "" {
		pageURL = fullURL // set url to fullURL
	}

	// HTTP GET request using newrequest for more flexibility
	req, err := http.NewRequest("GET", pageURL, nil) // GET request, so no response body

	// HTTP request check
	if err != nil {
		return LocationAreaResponse{}, fmt.Errorf("error with HTTP request: %w", err) // empty slice & error
	}

	// modify GET request header (not required, but BEST GO PRACTICE)
	req.Header.Set("Accept", "application/json") // expects json data as HTTP response
	// CORE: "Content-Type" - sending TO server, "Accept" - response FROM server

	// create new HTTP client
	//client := &http.Client{} // bad practice to use DefaultClient
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

	// can now return the location area response from server as success
	return locationRes, nil // nil error
}
