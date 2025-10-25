package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/admin/wordle/pkg/client"
)

func main() {
	// Command line flags
	serverURL := flag.String("server", "http://localhost:8080", "server URL")
	flag.Parse()

	// Create and run client application
	app := client.NewApp(*serverURL, os.Stdin)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
