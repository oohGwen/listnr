package events

import (
	"sync"
	"time"

	"github.com/sammwyy/listnr/internal/library"
)

type EventType string

const (
	SongChanged     EventType = "song_changed"
	PlaybackPaused  EventType = "playback_paused"
	PlaybackResumed EventType = "playback_resumed"
	SongEnded       EventType = "song_ended"
	ProgressUpdated EventType = "progress_updated"
	VolumeChanged   EventType = "volume_changed"
)

type Event struct {
	Type EventType
	Data interface{}
}

type SongData struct {
	Song *library.Song
}

type PlaybackData struct {
	IsPlaying bool
}

type ProgressData struct {
	Current time.Duration
	Total   time.Duration
	Song    *library.Song
}

type VolumeData struct {
	Level float64
}

type SongEndedData struct {
	Song *library.Song
}

type EventBus struct {
	subscribers map[EventType][]chan Event
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (eb *EventBus) Subscribe(eventType EventType) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 100)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if subscribers, exists := eb.subscribers[event.Type]; exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// Channel is full, skip this subscriber
			}
		}
	}
}
