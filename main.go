package main

import (
	"fmt"
	"log"
	"os"

	"github.com/initia-labs/weave/cmd"
)

func main() {
	if err := executeWithRecovery(); err != nil {
		log.Fatal(err)
	}
}

func executeWithRecovery() (err error) {
	defer func() {
		if r := recover(); r != nil {
			// Print a clean error message to stderr on panic
			fmt.Fprintln(os.Stderr, "An unexpected error occurred:", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	// Execute the main command
	return cmd.Execute()
}
