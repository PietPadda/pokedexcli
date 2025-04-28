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
	callback    func(*config) error // needs to accepet the *config pointer
}

// CORE: we pass config ptr to callback to allow NEXT & PREVIOUS pagination to all commands

// returns a map of all REPL commands we have
// all commands are "registered" here
func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": { // exit command
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": { // help command
			name:        "help",
			description: "List all Commands",
			callback:    commandHelp,
		},
		"map": { // map command
			name:        "map",
			description: "List next 20 Locations",
			callback:    commandMap,
		},
		"mapb": { // mapb command
			name:        "mapb",
			description: "List previous 20 Locations",
			callback:    commandMapb,
		},
	}
}

// callback - terminates the program
func commandExit(cfg *config) error { // accepts config file now for pagination
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0) // neatly terminate program
	return nil // no error on exit
}

// callback - lists all registered commands
func commandHelp(cfg *config) error { // accepts config file now for pagination
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
func commandMap(cfg *config) error { // accepts config file now for pagination
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
func commandMapb(cfg *config) error { // accepts config file now for pagination
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

		commandInput := cleanedInput[0] // get first word from input

		// loop through command list to check if input exists (registry lookup)
		command, ok := commands[commandInput] // see if input exists here

		// if it exists callback it
		if ok {
			err := command.callback(cfg) // return callback err value to car
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
