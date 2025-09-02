package main

import (
	"github.com/gdamore/tcell/v2"
)

// configures all keyboard shortcuts
func (a *App) SetupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			return a.handleEscapeKey()

		case tcell.KeyLeft:
			return a.handleLeftKey()

		case tcell.KeyRight:
			return a.handleRightKey()

		case tcell.KeyUp:
			return a.handleUpKey(event)

		case tcell.KeyDown:
			return a.handleDownKey(event)

		case tcell.KeyEnter, tcell.KeyRune:
			return a.handleRuneKey(event)
		}

		return event
	})
}

// handles ESC key press
func (a *App) handleEscapeKey() *tcell.EventKey {
	// Toggle between explorer and controls
	if a.focusMode == "explorer" {
		a.focusMode = "controls"
		// Don't change focus to controls widget, just change mode
	} else {
		a.focusMode = "explorer"
		a.app.SetFocus(a.sidebar)
	}
	a.UpdateControlsDisplay()
	return nil
}

// handles LEFT arrow key press
func (a *App) handleLeftKey() *tcell.EventKey {
	if a.focusMode == "explorer" {
		a.app.SetFocus(a.sidebar)
		return nil
	} else if a.focusMode == "controls" && a.player.currentSong != nil {
		a.SeekBackward()
		return nil
	}
	return nil
}

// handles RIGHT arrow key press
func (a *App) handleRightKey() *tcell.EventKey {
	if a.focusMode == "explorer" {
		a.app.SetFocus(a.songList)
		return nil
	} else if a.focusMode == "controls" && a.player.currentSong != nil {
		a.SeekForward()
		return nil
	}
	return nil
}

// handles UP arrow key press
func (a *App) handleUpKey(event *tcell.EventKey) *tcell.EventKey {
	if a.focusMode == "controls" {
		if a.player.currentSong != nil {
			a.VolumeUp()
		}
		return nil // Consume the event to prevent scrolling
	}
	return event // Let the normal navigation work in explorer mode
}

// handles DOWN arrow key press
func (a *App) handleDownKey(event *tcell.EventKey) *tcell.EventKey {
	if a.focusMode == "controls" {
		if a.player.currentSong != nil {
			a.VolumeDown()
		}
		return nil // Consume the event to prevent scrolling
	}
	return event // Let the normal navigation work in explorer mode
}

// handles character key presses
func (a *App) handleRuneKey(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case ' ':
			// Toggle play/pause
			a.TogglePlayPause()
			return nil
		case 'a', 'A':
			// Previous song
			a.PreviousSong()
			return nil
		case 'd', 'D':
			// Next song
			a.NextSong()
			return nil
		case 'q', 'Q':
			// Quit application
			a.app.Stop()
			return nil
		}
	}
	return event
}
