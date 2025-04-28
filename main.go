// main.go
package main // all files in same folder form part of package main

import (
	// import internal packages
	"github.com/PietPadda/pokedexcli/internal/pokeapi" // our internal package pokeapi
)

func main() {
	// create the pokeapi client
	pokeClient := pokeapi.NewClient()

	// call start REPL to run the application
	startREPL(pokeClient) // startrepl will use this for api requests
}
