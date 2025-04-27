// repl.go
package main // all files in same folder form part of package main

import (
	"bufio"   // for input blocking
	"fmt"     // for printing
	"os"      // for OS input
	"strings" // for Fields (split whitespace) and ToLower (lowercase)
)

// our command registry (abstraction)
// allows use to manage all commands we'll be adding
type cliCommand struct {
	name        string
	description string
	callback    func() error
}

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
	}
}

// callback - terminates the program
func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0) // neatly terminate program
	return nil // no error on exit
}

// callback - lists all registered commands
func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:\n") // newline at end for separation of command list

	commands := getCommands() // get all commands

	// loop thru all commands and print them
	for _, command := range commands {
		// print command name and description
		fmt.Printf("%s: %s\n", command.name, command.description)
	}

	// return success
	return nil
}

// startREPL starts the Read-Eval-Print-Loop for the Pokedex CLI
func startREPL() {
	// block until user input
	scanner := bufio.NewScanner(os.Stdin) // wait for input
	commands := getCommands()             // get all commands

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
			err := command.callback() // return callback err value to car

			// callback check
			if err != nil {
				fmt.Println(err) // Errorf doesn't work here, we don't have error output
			}
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
