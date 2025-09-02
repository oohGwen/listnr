package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// plays a given song
func (a *App) PlaySong(song *Song) {
	// Stop current playback
	a.stopPlayback()

	file, err := os.Open(song.Path)
	if err != nil {
		return
	}

	var streamer beep.StreamSeekCloser
	var format beep.Format

	ext := strings.ToLower(filepath.Ext(song.Path))
	switch ext {
	case ".mp3":
		streamer, format, err = mp3.Decode(file)
	case ".wav":
		streamer, format, err = wav.Decode(file)
	default:
		file.Close()
		return
	}

	if err != nil {
		file.Close()
		return
	}

	// Setup player
	a.player.streamer = streamer
	a.player.format = format
	a.player.currentSong = song
	a.player.ctrl = &beep.Ctrl{Streamer: streamer, Paused: false}
	a.player.volume = &effects.Volume{
		Streamer: a.player.ctrl,
		Base:     2,
		Volume:   a.volumeToDecibels(a.config.Volume),
		Silent:   false,
	}
	a.player.isPlaying = true

	speaker.Play(a.player.volume)
	a.UpdateControlsDisplay()

	// Start progress update ticker
	a.startProgressUpdater()
}

// starts a goroutine that updates the UI every second
func (a *App) startProgressUpdater() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if a.player.streamer == nil || !a.player.isPlaying {
				return
			}

			// Update the display
			a.app.QueueUpdateDraw(func() {
				a.UpdateControlsDisplay()
			})
		}
	}()
}

// stops the current playback
func (a *App) stopPlayback() {
	if a.player.streamer != nil {
		speaker.Clear()
		a.player.streamer.Close()
		a.player.streamer = nil
		a.player.currentSong = nil
		a.player.isPlaying = false
	}
}

// toggles between play and pause
func (a *App) TogglePlayPause() {
	if a.player.ctrl != nil {
		speaker.Lock()
		a.player.ctrl.Paused = !a.player.ctrl.Paused
		a.player.isPlaying = !a.player.ctrl.Paused
		speaker.Unlock()
		a.UpdateControlsDisplay()
	}
}

// seeks forward by 5 seconds
func (a *App) SeekForward() {
	if a.player.streamer != nil {
		speaker.Lock()
		current := a.player.streamer.Position()
		newPos := current + a.player.format.SampleRate.N(time.Second*5)
		a.player.streamer.Seek(newPos)
		speaker.Unlock()
	}
}

// seeks backward by 5 seconds
func (a *App) SeekBackward() {
	if a.player.streamer != nil {
		speaker.Lock()
		current := a.player.streamer.Position()
		newPos := current - a.player.format.SampleRate.N(time.Second*5)
		if newPos < 0 {
			newPos = 0
		}
		a.player.streamer.Seek(newPos)
		speaker.Unlock()
	}
}

// increases the volume
func (a *App) VolumeUp() {
	a.config.Volume = min(1.0, a.config.Volume+0.05)
	if a.player.volume != nil {
		speaker.Lock()
		a.player.volume.Volume = a.volumeToDecibels(a.config.Volume)
		speaker.Unlock()
	}
	a.UpdateControlsDisplay()
}

// decreases the volume
func (a *App) VolumeDown() {
	a.config.Volume = max(0.0, a.config.Volume-0.05)
	if a.player.volume != nil {
		speaker.Lock()
		a.player.volume.Volume = a.volumeToDecibels(a.config.Volume)
		speaker.Unlock()
	}
	a.UpdateControlsDisplay()
}

// converts volume percentage to decibels
func (a *App) volumeToDecibels(volume float64) float64 {
	if volume == 0 {
		return -10 // Silent
	}
	return (volume - 1) * 5 // Convert 0-1 to -5-0 dB range
}

// plays the next song in the current directory
func (a *App) NextSong() {
	if a.currentDir == nil || len(a.currentDir.Songs) == 0 {
		return
	}

	a.selectedSong = (a.selectedSong + 1) % len(a.currentDir.Songs)
	a.PlaySong(&a.currentDir.Songs[a.selectedSong])
	a.songList.SetCurrentItem(a.selectedSong)
}

// plays the previous song in the current directory
func (a *App) PreviousSong() {
	if a.currentDir == nil || len(a.currentDir.Songs) == 0 {
		return
	}

	a.selectedSong = (a.selectedSong - 1 + len(a.currentDir.Songs)) % len(a.currentDir.Songs)
	a.PlaySong(&a.currentDir.Songs[a.selectedSong])
	a.songList.SetCurrentItem(a.selectedSong)
}
