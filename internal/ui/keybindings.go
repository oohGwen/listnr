package ui

import (
	"listnr/internal/audio"

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
	switch kh.app.GetFocusMode() {
	case "explorer":
		return kh.handleExplorerKeys(event)
	case "controls":
		return kh.handleControlKeys(event)
	}
	return event
}

func (kh *KeyHandler) handleExplorerKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		kh.app.SetFocusMode("controls")
		return nil
	case tcell.KeyLeft:
		kh.app.FocusLeft()
		return nil
	case tcell.KeyRight:
		kh.app.FocusRight()
		return nil
	case tcell.KeyRune:
		return kh.handleGlobalKeys(event)
	}
	return event
}

func (kh *KeyHandler) handleControlKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		kh.app.SetFocusMode("explorer")
		kh.app.FocusLeft()
		return nil
	case tcell.KeyLeft:
		if kh.player.CurrentSong() != nil {
			kh.player.SeekBackward()
		}
		return nil
	case tcell.KeyRight:
		if kh.player.CurrentSong() != nil {
			kh.player.SeekForward()
		}
		return nil
	case tcell.KeyUp:
		if kh.player.CurrentSong() != nil {
			kh.player.VolumeUp()
		}
		return nil
	case tcell.KeyDown:
		if kh.player.CurrentSong() != nil {
			kh.player.VolumeDown()
		}
		return nil
	case tcell.KeyRune:
		return kh.handleGlobalKeys(event)
	}
	return nil
}

func (kh *KeyHandler) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case ' ':
		kh.player.TogglePlayPause()
		return nil
	case 'a', 'A':
		kh.app.PreviousSong()
		return nil
	case 'd', 'D':
		kh.app.NextSong()
		return nil
	case 'q', 'Q':
		kh.app.Stop()
		return nil
	case 'r', 'R':
		kh.app.ToggleRepeatMode()
		return nil
	case 'n', 'N':
		kh.app.ToggleAutoplay()
		return nil
	}

	return event
}
