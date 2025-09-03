package ui

import (
	"github.com/sammwyy/listnr/internal/audio"

	"github.com/gdamore/tcell/v2"
)

type KeyHandler struct {
	app    *App
	player *audio.Player
}

func NewKeyHandler(app *App, player *audio.Player) *KeyHandler {
	return &KeyHandler{
		app:    app,
		player: player,
	}
}

func (kh *KeyHandler) Setup() {
	kh.app.tviewApp.SetInputCapture(kh.handleKey)
}

func (kh *KeyHandler) handleKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyLeft:
		kh.app.FocusLeft()
		return nil
	case tcell.KeyRight:
		kh.app.FocusRight()
		return nil
	case tcell.KeyEsc:
		kh.app.Stop()
		return nil
	case tcell.KeyRune:
		return kh.handleGlobalKeys(event)
	}
	return event
}

func (kh *KeyHandler) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case ' ':
		kh.player.TogglePlayPause()
		return nil
	// Nexr/Prev song
	case 'q', 'Q':
		kh.app.PreviousSong()
		return nil
	case 'e', 'E':
		kh.app.NextSong()
		return nil
	// Modes
	case 'r', 'R':
		kh.app.ToggleRepeatMode()
		return nil
	case 'n', 'N':
		kh.app.ToggleAutoplay()
		return nil
	// Volume
	case 'w', 'W':
		kh.player.VolumeUp()
		return nil
	case 's', 'S':
		kh.player.VolumeDown()
		return nil
	}

	return event
}
