package main

import (
	"os"

	"github.com/aljrico/Google-Play-Console-CLI/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
