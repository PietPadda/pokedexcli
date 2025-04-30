// repl.go
package main // all files in same folder form part of package main

import (
	// import standard libraries
	"bufio"     // for input blocking
	"fmt"       // for printing
	"math/rand" // for catch probability
	"os"        // for OS input
	"strings"   // for Fields (split whitespace) and ToLower (lowercase)

	// import internal packages
	"github.com/PietPadda/pokedexcli/internal/pokeapi" // our internal package pokeapi
)

// for paginating through location areas
type config struct {
	NextURL       string           // next 20 areas (map command)
	PrevURL       string           // previous 20 areas (mapb command)
	PokeapiClient pokeapi.Client   // client to make API calls
	Pokedex       *pokeapi.Pokedex // for storing caught pokemon
}

// our command registry (abstraction)
// allows use to manage all commands we'll be adding
type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
	// callback *config pointer for pagination & pokeapi client
	// callback []string for command handling of parameters
}

// CORE: we pass config ptr to callback to allow NEXT & PREVIOUS pagination to all commands

// returns a map of all REPL commands we have
// all commands are "registered" here
func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": { // exit command -- exit program
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": { // help command -- shows callback commands
			name:        "help",
			description: "List all Commands",
			callback:    commandHelp,
		},
		"map": { // map command -- paginates locations
			name:        "map",
			description: "List next 20 Locations",
			callback:    commandMap,
		},
		"mapb": { // mapb command -- depaginates locations
			name:        "mapb",
			description: "List previous 20 Locations",
			callback:    commandMapb,
		},
		"explore": { // explore command -- shows pokemons at location
			name:        "explore",
			description: "List pokemon available at Location (takes location arg)",
			callback:    commandExplore,
		},
		"catch": { // catch command -- attempt to catch pokemon at location
			name:        "catch",
			description: "Try to catch Pokemon (takes pokemon arg)",
			callback:    commandCatch,
		},
		"inspect": { // inspect command -- attempt to list stats of a pokemon in the pokedex (if caught)
			name:        "inspect",
			description: "Lists stats of pokemon in pokedex (takes pokemon arg)",
			callback:    commandInspect,
		},
		"pokedex": { // pokedex command -- lists all pokemon in the pokedex
			name:        "pokedex",
			description: "Lists all pokemon caught in the pokedex",
			callback:    commandPokedex,
		},
	}
}

// callback - terminates the program
// accepts config file for pagination & pokeapi client
// accepts args for command parameters
func commandExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0) // neatly terminate program
	return nil // no error on exit
}

// callback - lists all registered commands
// accepts config file for pagination & pokeapi client
// accepts args for command parameters
func commandHelp(cfg *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println() // newline at end for separation of command list
	// CORE: GO IS DUMB! can't just add a \n... need to make a NEW blank fmt.Println or get UNIT TEST ERRORS!!

	commands := getCommands() // get all commands

	// loop thru all commands and print them
	for _, command := range commands {
		// print command name and description
		fmt.Printf("%s: %s\n", command.name, command.description)
	}

	// return success
	return nil
}

// callback - prints the map locations and increases the URL pagination
// accepts config file for pagination & pokeapi client
// accepts args for command parameters
func commandMap(cfg *config, args []string) error {
	// first set the default url if no request has been made
	url := cfg.NextURL

	// no request made check
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area" // default starting url
	}

	// next we make API request using the pokeapi client
	res, err := cfg.PokeapiClient.GetLocationAreas(url) // pass the url here

	// server response check
	if err != nil {
		return err // return error
	}

	// successful response, use it to update URL pagination in config
	// res Next & Previous nil check (safe dereffing)
	// Next & Previous: from LocationAreaResponse (LAR) in client.go
	if res.Next != nil {
		cfg.NextURL = *res.Next // use ptr because CAN be null!
	} else {
		cfg.NextURL = "" // this handles the NULL case (no next page)
	}

	if res.Previous != nil {
		cfg.PrevURL = *res.Previous // use ptr because CAN be null!
	} else {
		cfg.PrevURL = "" // this handles the NULL case (no previous page)
	}

	// loop thru response results and print all to terminal
	fmt.Println("Location Areas:")         // initial print before looping
	for _, location := range res.Results { // from LocationAreaResponse (LAR) in client.go
		// print command name and description
		fmt.Println("- ", location.Name) // from LocationArea (LA) in client.go
	}

	// return success
	return nil
}

// callback - prints the map locations and decreases the URL pagination
// accepts config file for pagination & pokeapi client
// accepts args for command parameters
func commandMapb(cfg *config, args []string) error {
	// first set the default url if no request has been made
	url := cfg.PrevURL

	// no request made check
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area" // default starting url
	}

	// next we make API request using the pokeapi client
	res, err := cfg.PokeapiClient.GetLocationAreas(url) // pass the url here

	// server response check
	if err != nil {
		return err // return error
	}

	// successful response, use it to update URL pagination in config
	// res Next & Previous nil check (safe dereffing)
	// Next & Previous: from LocationAreaResponse (LAR) in client.go
	if res.Next != nil {
		cfg.NextURL = *res.Next // use ptr because CAN be null!
	} else {
		cfg.NextURL = "" // this handles the NULL case (no next page)
	}

	if res.Previous != nil {
		cfg.PrevURL = *res.Previous // use ptr because CAN be null!
	} else {
		cfg.PrevURL = "" // this handles the NULL case (no previous page)
	}

	// loop thru response results and print all to terminal
	fmt.Println("Location Areas:")         // initial print before looping
	for _, location := range res.Results { // from LocationAreaResponse (LAR) in client.go
		// print command name and description
		fmt.Println("- ", location.Name) // from LocationArea (LA) in client.go
	}

	// return success
	return nil
}

// callback - prints pokemon available at location arg
// accepts config file for pagination & pokeapi client
// accepts args for command parameters
func commandExplore(cfg *config, args []string) error {
	// nil ptr check (Go Best Practice)
	if cfg == nil {
		return fmt.Errorf("error: config is nil") // early return custom error
	}

	// args check
	if len(args) == 0 { // no arg(s) provided
		return fmt.Errorf("error: explore must take location area name as argument") // early return custom error
	}

	// get location area name from args
	locationAreaName := args[0] // location area is first arg

	// use pokeapi client to fetch the pokemon from this location
	res, err := cfg.PokeapiClient.GetLocationArea(locationAreaName) // pass location name here
	// REVIEW: config holds client field, client fetches data with method called on it, method uses location area

	// fetch check
	if err != nil {
		return fmt.Errorf("error client fetching pokemon from location: %w", err)
	}

	// loop thru response results and print all pokemon to terminal
	fmt.Printf("Exploring %s...\n", locationAreaName) // initial print before looping

	// no pokemon found check
	if len(res.PokemonEncounters) == 0 {
		fmt.Println("No Pokemon were found at this location.")
		return nil // still a success, just empty location
	}

	fmt.Println("Found Pokemon:")                     // initial print before looping
	for _, encounter := range res.PokemonEncounters { // from PokemonEncounters (PE) in client.go
		// print each pokemon with a newline
		fmt.Printf("- %s\n", encounter.Pokemon.Name) // from PokemonEncounters
	}

	// return success
	return nil
}

// callback - fetches pokemon details and attempts to catch it
// accepts config file for pagination & pokeapi client
// accepts args for command parameters
func commandCatch(cfg *config, args []string) error {
	// nil ptr check (Go Best Practice)
	if cfg == nil {
		return fmt.Errorf("error: config is nil") // early return custom error
	}

	// args check
	if len(args) == 0 { // no arg(s) provided
		return fmt.Errorf("error: catch must take pokemon name as argument") // early return custom error
	}

	// get location area name from args
	pokemonName := args[0] // pokemon name is first arg

	// use pokeapi client to fetch the pokemon details
	res, err := cfg.PokeapiClient.GetPokemonStats(pokemonName) // pass pokemon name here
	// REVIEW: config holds client field, client fetches data with method called on it, method uses location area

	// fetch check
	if err != nil {
		return fmt.Errorf("error client fetching pokemon details: %w", err)
	}

	// get pokemon base experience for catch probability calc
	baseExperience := res.BaseExperience // from PokemonStats struct

	// calculate probability of catching using inverse proportion for simplicity
	catchRate := 100.0 / (1.0 + float64(baseExperience)/60.0) // roll to beat, decreases with more base xp
	// 60xp = 50%
	// 180xp = 25%
	// 1140xp = 5%

	// determine catch success
	catchRoll := rand.Intn(96)                 // random roll from 0 to 95 (last int not incl)
	catchSuccess := catchRoll < int(catchRate) // if we roll less than catch rate, this is true ie caught

	// initial print before determining success or failure of ctaching
	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	// catch success check
	if catchSuccess { // true
		fmt.Printf("%s was caught!\n", pokemonName) // caught a pokemon

		// add to pokedex here -- to be implemented later
		cfg.Pokedex.PokemonAdd(pokemonName, res)
		// we use the method PokemonAdd on the pokedex to add a pokemon to it
		// Pokedex is init in config and thus a field of cfg

		// NOTE: res = PokemonStats!
		fmt.Printf("%s has been added to the Pokedex!\n", pokemonName) // indicate added to pokedex

	} else { // false
		fmt.Printf("%s escaped!\n", pokemonName) // it escaped
		// no pokemon added as catchSuccess is false
	}

	// return success
	return nil
}

// callback - prints pokemon stats that's caught in pokedex
// accepts config file for pokedex
// accepts args for command parameters
func commandInspect(cfg *config, args []string) error {
	// nil ptr check (Go Best Practice)
	if cfg == nil {
		return fmt.Errorf("error: config is nil") // early return custom error
	}

	// args check
	if len(args) == 0 { // no arg(s) provided
		return fmt.Errorf("error: inspect must take pokemon name as argument") // early return custom error
	}

	// get location area name from args
	pokemonName := args[0] // pokemon name is first arg

	// NOTE: don't need use pokeapi client to fetch as it's already caught (supposed to be) and in pokedex!

	// loop through pokedex to check if pokemon exists (comma-ok check + bonus err)
	pokemon, ok, err := cfg.Pokedex.PokemonGet(pokemonName) // see if input exists here

	// pokedex entries call check
	if err != nil {
		return fmt.Errorf("error getting pokedex entry: %w", err) // only err message
	}

	// pokemon found check
	if !ok { //if ok return false
		fmt.Println("you have not caught that pokemon") // display not found in pokedex to user
		return nil                                      // return success
	}

	// display the pokemon's stats
	fmt.Printf("Name: %s\n", pokemon.Name)     // display name
	fmt.Printf("Height: %d\n", pokemon.Height) // display height
	fmt.Printf("Weight: %d\n", pokemon.Weight) // display weight

	// display stats header before looping
	fmt.Println("Stats:")
	// loop thru stats and display them with names
	for _, stat := range pokemon.Stats { // stats contains name and basestat value
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat) // print each name and int value
	}
	// PokemonStats struct contains a PokemonStat struct with array "Stat", which has a field "Name" thus stat.Stat.Name
	// PokemonStats contains a PokemonStat struct with field "BaseState" thus stat.BaseStat

	// display types header before looping
	fmt.Println("Types:")
	// loop thru stats and display them with names
	for _, pokemonType := range pokemon.Types { // types array just has names
		fmt.Printf("  - %s\n", pokemonType.Type.Name) // print each type
	}
	// PokemonStats struct contains a PokemonTypes struct with array "Type", which has a field "Name" thus pokemonType.Type.Name

	// return success
	return nil
}

// callback - prints all pokemon caught in the pokedex
// accepts config file for pokedex
func commandPokedex(cfg *config, args []string) error {
	// nil ptr check (Go Best Practice)
	if cfg == nil {
		return fmt.Errorf("error: config is nil") // early return custom error
	}

	// NOTE: don't need use pokeapi client to fetch as it's already caught (supposed to be) and in pokedex!

	// display pokedex header before looping
	fmt.Println("Your Pokedex:")

	// get all the pokemon from pokedex
	names, err := cfg.Pokedex.PokemonGetAllCaught() // save all names and err to vars
	// apply method to pokedex which is init in config

	// get all names check
	if err != nil {
		return fmt.Errorf("error getting all pokemon from pokedex: %w", err) // early return
	}

	// empty amp check
	if len(names) == 0 {
		fmt.Println("You have not caught any pokemon yet!")
		return nil //early return
	}

	// there are pokemon! let's proceed with print

	// loop thru pokedex to get names
	for _, pokemonName := range names { // names of pokemon in pokedex
		fmt.Printf(" - %s\n", pokemonName) // print pokedex pokemon
	}

	// return success
	return nil
}

// startREPL starts the Read-Eval-Print-Loop for the Pokedex CLI
func startREPL(pokeClient pokeapi.Client) {
	// block until user input
	scanner := bufio.NewScanner(os.Stdin) // wait for input
	commands := getCommands()             // get all commands
	cfg := &config{
		PokeapiClient: pokeClient,           // store client in config
		Pokedex:       pokeapi.NewPokedex(), // store pokedex in config,
	} // init config ptr for NEXT & PREVIOUS pagination

	// infinite loop
	for {
		// print our CLI prompt
		fmt.Print("Pokedex > ") // no newline

		// get user input and clean
		scanner.Scan()                        // read the user input
		userInput := scanner.Text()           // get the user input
		cleanedInput := cleanInput(userInput) // clean the input (lowercase,no WS, slice)

		// empty input edge case handle
		if len(cleanedInput) == 0 {
			continue // don't do the rest
		}

		commandInput := cleanedInput[0] // get command (first word) from input
		args := cleanedInput[1:]        // get args (rest of words) from input

		// loop through command list to check if input exists (registry lookup)
		command, ok := commands[commandInput] // see if input exists here

		// if it exists callback it
		if ok {
			err := command.callback(cfg, args) // return callback err value to var
			// CORE: need to pass config file here to call funcs to allow pagination

			// callback check
			if err != nil {
				fmt.Println(err) // Errorf doesn't work here, we don't have error output
			}
			continue // Skip the "Invalid command" message when command is valid
		}

		// otherwise it doesn't exist
		fmt.Println("Invalid command")
	}
}

// make the words lower cases + plit words by whitespace (removes WS)
func cleanInput(text string) []string {
	// empty str edge case
	if len(text) == 0 {
		return []string{} // return empty slice
	}

	// make it lowercase
	lowerTxt := strings.ToLower(text) // .ToLower() does this

	// we handle spaces by splitting and making new slice
	splitTxt := strings.Fields(lowerTxt) // .Fields() does this
	return splitTxt                      // slice of cleaned strings
}
