// repl.go
package main // all files in same folder form part of package main

import "strings" // for Fields (split whitespace) and ToLower (lowercase)

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
