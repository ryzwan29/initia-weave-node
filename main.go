package main

import (
	"log"

	"github.com/initia-labs/weave/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
