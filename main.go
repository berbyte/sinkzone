package main

import (
	"fmt"
	"os"

	"github.com/berbyte/sinkzone/cmd"
	_ "modernc.org/sqlite"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
