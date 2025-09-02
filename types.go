package main

import (
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/rivo/tview"
)

// Config structure for the application
type Config struct {
	MusicRoutes []string `json:"music_routes"`
	Volume      float64  `json:"volume"`
	LastPath    string   `json:"last_path"`
}

// Song represents a music file
type Song struct {
	Path     string
	Name     string
	Duration time.Duration
}

// Directory represents a directory with songs
type Directory struct {
	Path  string
	Name  string
	Songs []Song
	Dirs  []Directory
}

// Player handles music playback
type Player struct {
	streamer    beep.StreamSeekCloser
	ctrl        *beep.Ctrl
	volume      *effects.Volume
	format      beep.Format
	currentSong *Song
	isPlaying   bool
}

// App holds the main application state
type App struct {
	app          *tview.Application
	config       *Config
	player       *Player
	directories  []Directory
	currentDir   *Directory
	selectedSong int
	selectedDir  int
	focusMode    string // "explorer" or "controls"

	// UI components
	sidebar  *tview.List
	songList *tview.List
	controls *tview.TextView
	layout   *tview.Flex
}

// Audio file extensions supported
var supportedExts = map[string]bool{
	".mp3": true,
	".wav": true,
}
