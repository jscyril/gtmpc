// cmd/client — main.go is the entry point for the gtmpc TUI client.
// It connects to a running gtmpc server, authenticates the user, and provides
// a terminal-based music player interface.
//
// Usage:
//
//	gtmpc-client [--server http://localhost:8080]
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jscyril/golang_music_player/internal/audio"
	"github.com/jscyril/golang_music_player/internal/ui"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	serverURL := flag.String("server", "http://localhost:8080", "Base URL of the gtmpc server")
	flag.Parse()

	// Initialise the API client
	client := apiclient.NewAPIClient(*serverURL)

	// Initialise the audio engine (speaker) and start it
	ctx := context.Background()
	engine := audio.NewAudioEngine()
	if err := engine.Start(ctx); err != nil {
		return fmt.Errorf("audio engine start: %w", err)
	}

	// Run the client-mode TUI
	if err := ui.RunClientApp(client, engine); err != nil {
		return fmt.Errorf("ui: %w", err)
	}

	return nil
}
