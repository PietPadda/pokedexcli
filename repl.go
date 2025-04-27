// repl.go
package main // all files in same folder form part of package main

import (
	"bufio"   // for input blocking
	"fmt"     // for printing
	"os"      // for OS input
	"strings" // for Fields (split whitespace) and ToLower (lowercase)
)

// we'll call this in main
func startREPL() {
	// block until user input
	scanner := bufio.NewScanner(os.Stdin) // wait for input

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

		firstWord := cleanedInput[0] // get first word from input

		// print first word to CLI
		fmt.Printf("Your command was: %s\n", firstWord)
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
