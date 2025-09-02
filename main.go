package main

import (
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

func main() {
	// Initialize speaker for audio playback
	speaker.Init(beep.SampleRate(44100), beep.SampleRate(44100).N(time.Second/10))

	app := NewApp()

	// Load configuration
	app.LoadConfig()

	// Scan music directories
	app.ScanDirectories()

	// Setup UI
	app.SetupUI()

	// Setup key bindings
	app.SetupKeyBindings()

	// Start the application
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
