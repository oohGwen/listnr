package audio

import (
	"context"
	"sync"
	"time"

	"github.com/sammwyy/listnr/internal/events"
	"github.com/sammwyy/listnr/internal/library"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
)

type Player struct {
	// Audio components
	sampleRate beep.SampleRate
	streamer   beep.StreamSeekCloser
	analyzer   *AudioAnalyzer
	ctrl       *beep.Ctrl
	volume     *effects.Volume
	format     beep.Format

	// State
	currentSong *library.Song
	isPlaying   bool
	volumeLevel float64

	// Communication
	eventBus *events.EventBus
	commands chan Command

	// Synchronization
	mu sync.RWMutex
}

type Command struct {
	Type string
	Args interface{}
}

const (
	CmdPlay     = "play"
	CmdPause    = "pause"
	CmdStop     = "stop"
	CmdSeek     = "seek"
	CmdVolume   = "volume"
	CmdNext     = "next"
	CmdPrevious = "previous"
)

func NewPlayer(sampleRate beep.SampleRate) *Player {
	return &Player{
		eventBus:    events.NewEventBus(),
		commands:    make(chan Command, 10),
		volumeLevel: 0.5,
		isPlaying:   false,
		sampleRate:  sampleRate,
	}
}

func (p *Player) Start(ctx context.Context) {
	go p.processCommands(ctx)
	go p.updateProgress(ctx)
}

func (p *Player) processCommands(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-p.commands:
			switch cmd.Type {
			case CmdPlay:
				if song, ok := cmd.Args.(*library.Song); ok {
					p.play(song)
				}
			case CmdPause:
				p.togglePlayPause()
			case CmdStop:
				p.stop()
			case CmdSeek:
				if duration, ok := cmd.Args.(time.Duration); ok {
					p.seek(duration)
				}
			case CmdVolume:
				if level, ok := cmd.Args.(float64); ok {
					p.setVolume(level)
				}
			}
		}
	}
}

func (p *Player) updateProgress(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.mu.RLock()
			if p.streamer != nil && p.isPlaying {
				position := p.streamer.Position()
				total := p.streamer.Len()

				if position >= total-1000 {
					p.eventBus.Publish(events.Event{
						Type: events.SongEnded,
						Data: events.SongEndedData{Song: p.currentSong},
					})
				}

				currentTime := time.Duration(position) * time.Second / time.Duration(p.format.SampleRate)
				totalTime := time.Duration(total) * time.Second / time.Duration(p.format.SampleRate)

				p.eventBus.Publish(events.Event{
					Type: events.ProgressUpdated,
					Data: events.ProgressData{
						Current: currentTime,
						Total:   totalTime,
						Song:    p.currentSong,
					},
				})
			}
			p.mu.RUnlock()
		}
	}
}

func (p *Player) play(song *library.Song) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop current playback
	p.stopInternal()

	// Decode new file
	streamer, format, err := DecodeFile(song.Path)
	if err != nil {
		return
	}

	// Resample if needed
	var rs beep.Streamer
	if format.SampleRate != p.sampleRate {
		rs = beep.Resample(3, format.SampleRate, p.sampleRate, streamer)
	} else {
		rs = streamer
	}

	// Setup audio analyzer
	p.analyzer = NewAudioAnalyzer(rs, p.eventBus)

	// Setup player components
	p.streamer = streamer
	p.format = format
	p.currentSong = song
	p.ctrl = &beep.Ctrl{Streamer: p.analyzer, Paused: false}
	p.volume = &effects.Volume{
		Streamer: p.ctrl,
		Base:     2,
		Volume:   p.volumeToDecibels(p.volumeLevel),
		Silent:   false,
	}
	p.isPlaying = true

	speaker.Play(p.volume)

	// Notify UI
	p.eventBus.Publish(events.Event{
		Type: events.SongChanged,
		Data: events.SongData{Song: song},
	})
	p.eventBus.Publish(events.Event{
		Type: events.PlaybackResumed,
		Data: events.PlaybackData{IsPlaying: true},
	})
}

func (p *Player) togglePlayPause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = !p.ctrl.Paused
		p.isPlaying = !p.ctrl.Paused
		speaker.Unlock()

		if p.isPlaying {
			p.eventBus.Publish(events.Event{
				Type: events.PlaybackResumed,
				Data: events.PlaybackData{IsPlaying: true},
			})
		} else {
			p.eventBus.Publish(events.Event{
				Type: events.PlaybackPaused,
				Data: events.PlaybackData{IsPlaying: false},
			})
		}
	}
}

func (p *Player) stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopInternal()
}

func (p *Player) stopInternal() {
	if p.streamer != nil {
		speaker.Clear()
		p.streamer.Close()
		p.streamer = nil
		p.currentSong = nil
		p.isPlaying = false
		p.analyzer = nil

		p.eventBus.Publish(events.Event{
			Type: events.PlaybackPaused,
			Data: events.PlaybackData{IsPlaying: false},
		})
	}
}

func (p *Player) seek(offset time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer != nil {
		speaker.Lock()
		current := p.streamer.Position()
		newPos := current + int(offset.Seconds()*float64(p.format.SampleRate))

		if newPos < 0 {
			newPos = 0
		}
		if newPos >= p.streamer.Len() {
			newPos = p.streamer.Len() - 1
		}

		p.streamer.Seek(newPos)
		speaker.Unlock()
	}
}

func (p *Player) setVolume(level float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.volumeLevel = clamp(level, 0.0, 1.0)

	if p.volume != nil {
		speaker.Lock()
		p.volume.Volume = p.volumeToDecibels(p.volumeLevel)
		speaker.Unlock()
	}

	p.eventBus.Publish(events.Event{
		Type: events.VolumeChanged,
		Data: events.VolumeData{Level: p.volumeLevel},
	})
}

func (p *Player) volumeToDecibels(volume float64) float64 {
	if volume == 0 {
		return -10 // Silent
	}
	return (volume - 1) * 5 // Convert 0-1 to -5-0 dB range
}

// Public API methods
func (p *Player) Play(song *library.Song) {
	p.commands <- Command{Type: CmdPlay, Args: song}
}

func (p *Player) TogglePlayPause() {
	p.commands <- Command{Type: CmdPause}
}

func (p *Player) Stop() {
	p.commands <- Command{Type: CmdStop}
}

func (p *Player) SeekForward() {
	p.commands <- Command{Type: CmdSeek, Args: 5 * time.Second}
}

func (p *Player) SeekBackward() {
	p.commands <- Command{Type: CmdSeek, Args: -5 * time.Second}
}

func (p *Player) VolumeUp() {
	p.mu.RLock()
	newVolume := p.volumeLevel + 0.05
	p.mu.RUnlock()
	p.commands <- Command{Type: CmdVolume, Args: newVolume}
}

func (p *Player) VolumeDown() {
	p.mu.RLock()
	newVolume := p.volumeLevel - 0.05
	p.mu.RUnlock()
	p.commands <- Command{Type: CmdVolume, Args: newVolume}
}

func (p *Player) EventBus() *events.EventBus {
	return p.eventBus
}

func (p *Player) CurrentSong() *library.Song {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentSong
}

func (p *Player) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isPlaying
}

func (p *Player) Volume() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.volumeLevel
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
