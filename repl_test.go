package main

import "testing" // importing testing package for unit tests

func TestCleanInput(t *testing.T) {
	// our unit tests
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ", // whitespace test
			expected: []string{"hello", "world"},
		},
		{
			input:    "Hello World", // lowercase test
			expected: []string{"hello", "world"},
		},
		{
			input:    "", // empty string test
			expected: []string{},
		},
		{
			input:    " hi  THERE, this IS a TeSt!", // empty string test
			expected: []string{"hi", "there,", "this", "is", "a", "test!"},
		},
	}

	// loop over cases to run all the tests
	for _, c := range cases { // unit test c
		// get result fomr cleanInput
		actual := cleanInput(c.input) // unit test's input

		// Check the length of the actual slice against the expected slice
		if len(actual) != len(c.expected) { // unit test's output
			t.Errorf("length of actual: '%v' and expected: '%v' slices do not match!",
				len(actual), len(c.expected))
			continue // stop the loop here -- one error at a time instead of being flooded!
		}
		// CORE: t.Errorf (not fmt) is for unit tests AND fails the tests!

		// loop thru slices and compare each word therein
		for i := range actual {
			resultWord := actual[i]
			expectedWord := c.expected[i]
			// Check each word in the slice
			if resultWord != expectedWord {
				t.Errorf("actual word: '%s' and expected word: '%s' do not match!",
					resultWord, expectedWord)
				continue // stop the loop here -- one error at a time instead of being flooded!
			}
		}
	}
}
