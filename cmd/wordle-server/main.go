package main

import (
	"flag"
	"log"

	"github.com/admin/wordle/internal/config"
	"github.com/admin/wordle/pkg/server"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "cfg/config.yaml", "path to configuration file")
	wordsPath := flag.String("words", "", "path to words list file (overrides config word_list)")
	port := flag.String("port", "8080", "server port")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Printf("Warning: failed to load config: %v, using defaults", err)
		cfg = config.DefaultConfig()
	}

	// If words file is specified, load words from file and override config word_list
	if *wordsPath != "" {
		words, err := config.LoadWordsFromFile(*wordsPath)
		if err != nil {
			log.Fatalf("Failed to load words file: %v", err)
		}
		cfg.WordList = words
		log.Printf("Loaded %d words from %s", len(words), *wordsPath)
	}

	// Create and start server application
	app := server.NewApp(cfg, *port)
	if err := app.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
