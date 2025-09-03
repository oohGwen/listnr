package ui

import (
	"context"
	"sync"

	"listnr/internal/audio"
	"listnr/internal/config"
	"listnr/internal/events"
	"listnr/internal/library"
	"listnr/internal/ui/components"

	"github.com/rivo/tview"
)

type App struct {
	// Core components
	tviewApp *tview.Application
	player   *audio.Player
	library  *library.Library
	config   *config.Config

	// UI state
	selectedSong    int
	currentDir      *library.Directory
	autoplayEnabled bool
	repeatMode      bool

	// UI components
	sidebar  *components.Sidebar
	songList *components.SongList
	controls *components.Controls
	layout   *tview.Flex

	// Event handling
	ctx        context.Context
	cancel     context.CancelFunc
	keyHandler *KeyHandler

	// Synchronization
	mu sync.RWMutex
}

func NewApp(cfg *config.Config, player *audio.Player, lib *library.Library) *App {
	app := &App{
		tviewApp:        tview.NewApplication(),
		player:          player,
		library:         lib,
		selectedSong:    0,
		autoplayEnabled: true,
		repeatMode:      false,
		config:          cfg,
	}

	app.keyHandler = NewKeyHandler(app, player)
	return app
}

func (a *App) Start(ctx context.Context) error {
	a.ctx, a.cancel = context.WithCancel(ctx)
	defer a.cancel()

	// Start audio player
	a.player.Start(a.ctx)

	// Setup UI
	a.setupUI()

	// Setup event handlers
	go a.handlePlayerEvents()

	// Setup keybindings
	a.keyHandler.Setup()

	// Start TUI
	return a.tviewApp.SetRoot(a.layout, true).EnableMouse(true).Run()
}

func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	if a.tviewApp != nil {
		a.tviewApp.Stop()
	}
}

func (a *App) setupUI() {
	// Initialize components
	a.sidebar = components.NewSidebar()
	a.songList = components.NewSongList()
	a.controls = components.NewControls()

	// Sync data
	a.controls.SetAutoplay(a.autoplayEnabled)
	a.controls.SetRepeatMode(a.repeatMode)

	// Setup component callbacks
	a.sidebar.SetSelectionCallback(a.onDirectorySelected)
	a.songList.SetSelectionCallback(a.onSongSelected)

	// Populate data
	a.populateLibrary()

	// Layout setup
	topLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.sidebar.List, 0, 1, true).
		AddItem(a.songList.List, 0, 3, false)

	a.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topLayout, 0, 1, true).
		AddItem(a.controls.TextView, 4, 0, false)

	// Set initial focus
	a.FocusLeft()
}

func (a *App) populateLibrary() {
	directories := a.library.GetDirectories()
	a.sidebar.SetDirectories(directories)

	if len(directories) > 0 {
		a.currentDir = directories[0]
		a.songList.SetDirectory(a.currentDir)
	}
}

func (a *App) onDirectorySelected(dir *library.Directory) {
	a.mu.Lock()
	a.currentDir = dir
	a.selectedSong = 0
	a.mu.Unlock()

	a.songList.SetDirectory(dir)
}

func (a *App) onSongSelected(song *library.Song, index int) {
	a.mu.Lock()
	a.selectedSong = index
	a.mu.Unlock()

	a.player.Play(song)
}

func (a *App) handlePlayerEvents() {
	progressCh := a.player.EventBus().Subscribe(events.ProgressUpdated)
	songCh := a.player.EventBus().Subscribe(events.SongChanged)
	playbackCh := a.player.EventBus().Subscribe(events.PlaybackResumed)
	pauseCh := a.player.EventBus().Subscribe(events.PlaybackPaused)
	volumeCh := a.player.EventBus().Subscribe(events.VolumeChanged)
	songEndedCh := a.player.EventBus().Subscribe(events.SongEnded)

	for {
		select {
		case <-a.ctx.Done():
			return
		case event := <-progressCh:
			if data, ok := event.Data.(events.ProgressData); ok {
				a.tviewApp.QueueUpdateDraw(func() {
					a.controls.UpdateProgress(data.Current, data.Total)
				})
			}
		case event := <-songCh:
			if data, ok := event.Data.(events.SongData); ok {
				a.tviewApp.QueueUpdateDraw(func() {
					a.controls.SetCurrentSong(data.Song)
				})
			}
		case event := <-playbackCh:
			if data, ok := event.Data.(events.PlaybackData); ok {
				a.tviewApp.QueueUpdateDraw(func() {
					a.controls.SetPlaying(data.IsPlaying)
				})
			}
		case event := <-pauseCh:
			if data, ok := event.Data.(events.PlaybackData); ok {
				a.tviewApp.QueueUpdateDraw(func() {
					a.controls.SetPlaying(data.IsPlaying)
				})
			}
		case event := <-volumeCh:
			if data, ok := event.Data.(events.VolumeData); ok {
				a.tviewApp.QueueUpdateDraw(func() {
					a.controls.SetVolume(data.Level)
				})
			}
		case event := <-songEndedCh:
			if data, ok := event.Data.(events.SongEndedData); ok {
				a.handleSongEnded(data.Song)
			}
		}
	}
}

func (a *App) handleSongEnded(song *library.Song) {
	a.mu.RLock()
	autoplay := a.autoplayEnabled
	repeat := a.repeatMode
	a.mu.RUnlock()

	if repeat {
		// Repeat if repeat is enabled
		a.player.Play(song)
	} else if autoplay {
		// Next song if autoplay is enabled
		a.NextSong()
	}
}

// Navigation methods
func (a *App) NextSong() {
	a.mu.RLock()
	if a.currentDir == nil || len(a.currentDir.Songs) == 0 {
		a.mu.RUnlock()
		return
	}

	newIndex := (a.selectedSong + 1) % len(a.currentDir.Songs)
	song := a.currentDir.Songs[newIndex]
	a.mu.RUnlock()

	a.mu.Lock()
	a.selectedSong = newIndex
	a.mu.Unlock()

	a.player.Play(song)
	a.songList.SetCurrentItem(newIndex)
}

func (a *App) PreviousSong() {
	a.mu.RLock()
	if a.currentDir == nil || len(a.currentDir.Songs) == 0 {
		a.mu.RUnlock()
		return
	}

	newIndex := (a.selectedSong - 1 + len(a.currentDir.Songs)) % len(a.currentDir.Songs)
	song := a.currentDir.Songs[newIndex]
	a.mu.RUnlock()

	a.mu.Lock()
	a.selectedSong = newIndex
	a.mu.Unlock()

	a.player.Play(song)
	a.songList.SetCurrentItem(newIndex)
}

func (a *App) FocusLeft() {
	a.tviewApp.SetFocus(a.sidebar.List)
	a.sidebar.SetFocused(true)
	a.songList.SetFocused(false)
}

func (a *App) FocusRight() {
	a.tviewApp.SetFocus(a.songList.List)
	a.sidebar.SetFocused(false)
	a.songList.SetFocused(true)
}

func (a *App) GetTviewApp() *tview.Application {
	return a.tviewApp
}

// State management
func (a *App) ToggleRepeatMode() {
	a.mu.Lock()
	a.repeatMode = !a.repeatMode
	a.mu.Unlock()
	a.controls.SetRepeatMode(a.repeatMode)
}

func (a *App) ToggleAutoplay() {
	a.mu.Lock()
	a.autoplayEnabled = !a.autoplayEnabled
	a.mu.Unlock()
	a.controls.SetAutoplay(a.autoplayEnabled)
}
