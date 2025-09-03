package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"listnr/internal/audio"
	"listnr/internal/config"
	"listnr/internal/library"
	"listnr/internal/ui"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

func main() {
	// Initialize speaker for audio playback
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down...")
		cancel()
	}()

	// Initialize components
	player := audio.NewPlayer(sampleRate)
	lib := library.NewLibrary()

	// Scan music directories
	if err := lib.Scan(cfg.MusicRoutes); err != nil {
		log.Fatal("Failed to scan music directories:", err)
	}

	// Create and start UI
	app := ui.NewApp(cfg, player, lib)
	if err := app.Start(ctx); err != nil {
		log.Fatal("Application error:", err)
	}
}
