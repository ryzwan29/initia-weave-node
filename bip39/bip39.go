package bip39

import (
	"strings"
)

// IsMnemonicValid attempts to verify that the provided mnemonic is valid.
// Validity is determined by both the number of words being appropriate,
// and that all the words in the mnemonic are present in the word list.
func IsMnemonicValid(mnemonic string) bool {
	// Create a list of all the words in the mnemonic sentence
	words := strings.Fields(mnemonic)

	//Get num of words
	numOfWords := len(words)

	// The number of words should be 12, 15, 18, 21 or 24
	if numOfWords < 12 || numOfWords > 24 || numOfWords%3 != 0 {
		return false
	}

	// Check if all words belong in the wordlist
	for i := 0; i < numOfWords; i++ {
		if _, ok := ReverseWordMap[words[i]]; !ok {
			return false
		}
	}

	return true
}
