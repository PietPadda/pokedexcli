// main.go
package main // all files in same folder form part of package main

import (
	// import standard Go libraries
	"time" // for interval limit pass to cache

	// import internal packages
	"github.com/PietPadda/pokedexcli/internal/pokeapi"   // pokeapi client package
	"github.com/PietPadda/pokedexcli/internal/pokecache" // cache package
)

func main() {
	// create cache for performant results
	cache := pokecache.NewCache(5 * time.Minute) // set cache to 2 minutes

	// create the pokeapi client
	pokeClient := pokeapi.NewClient(cache)

	// call start REPL to run the application
	startREPL(pokeClient) // startrepl will use this for api requests
}
