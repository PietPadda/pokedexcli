// repl.go
package main // all files in same folder form part of package main

import (
	// import standard libraries
	"bufio"   // for input blocking
	"fmt"     // for printing
	"os"      // for OS input
	"strings" // for Fields (split whitespace) and ToLower (lowercase)

	// import internal packages
	"github.com/PietPadda/pokedexcli/internal/pokeapi" // our internal package pokeapi
)

// for paginating through location areas
type config struct {
	NextURL       string         // next 20 areas (map command)
	PrevURL       string         // previous 20 areas (mapb command)
	PokeapiClient pokeapi.Client // client to make API calls
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

// startREPL starts the Read-Eval-Print-Loop for the Pokedex CLI
func startREPL(pokeClient pokeapi.Client) {
	// block until user input
	scanner := bufio.NewScanner(os.Stdin) // wait for input
	commands := getCommands()             // get all commands
	cfg := &config{
		PokeapiClient: pokeClient, // store client in config
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
