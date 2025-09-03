package audio

import (
	"math"
	"time"

	"github.com/sammwyy/listnr/internal/events"

	"github.com/gopxl/beep"
)

type AudioAnalyzer struct {
	beep.Streamer
	samples      []float64
	sampleBuffer [][]float64
	bufferSize   int
	bufferPos    int
	eventBus     *events.EventBus
	lastUpdate   time.Time
}

func NewAudioAnalyzer(streamer beep.Streamer, eventBus *events.EventBus) *AudioAnalyzer {
	return &AudioAnalyzer{
		Streamer:     streamer,
		samples:      make([]float64, 1024),
		sampleBuffer: make([][]float64, 2), // Stereo
		bufferSize:   512,
		bufferPos:    0,
		eventBus:     eventBus,
		lastUpdate:   time.Now(),
	}
}

func (a *AudioAnalyzer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = a.Streamer.Stream(samples)

	if ok && n > 0 {
		a.analyzeSamples(samples[:n])
	}

	return n, ok
}

func (a *AudioAnalyzer) analyzeSamples(samples [][2]float64) {
	// Throttle updates
	now := time.Now()
	if now.Sub(a.lastUpdate) < 50*time.Millisecond {
		return
	}
	a.lastUpdate = now

	var totalAmplitude float64
	for _, sample := range samples {
		left := math.Abs(sample[0])
		right := math.Abs(sample[1])
		amplitude := (left + right) / 2
		totalAmplitude += amplitude
	}

	if len(samples) > 0 {
		totalAmplitude /= float64(len(samples))
	}

	// Simulate brand freq (To-Do: switch to a FFT system)
	frequencyBands := a.simulateFrequencyBands(samples, totalAmplitude)

	// Publish event
	a.eventBus.Publish(events.Event{
		Type: events.AudioDataUpdated,
		Data: events.AudioData{
			FrequencyBands: frequencyBands,
			Amplitude:      totalAmplitude,
			IsPlaying:      true,
		},
	})
}

func (a *AudioAnalyzer) simulateFrequencyBands(samples [][2]float64, baseAmplitude float64) []float64 {
	bands := make([]float64, 16)

	if len(samples) == 0 {
		return bands
	}

	samplesPerBand := len(samples) / 16
	if samplesPerBand < 1 {
		samplesPerBand = 1
	}

	for i := 0; i < 16; i++ {
		var bandAmplitude float64
		start := i * samplesPerBand
		end := start + samplesPerBand

		if end > len(samples) {
			end = len(samples)
		}

		for j := start; j < end; j++ {
			left := math.Abs(samples[j][0])
			right := math.Abs(samples[j][1])
			bandAmplitude += (left + right) / 2
		}

		if end > start {
			bandAmplitude /= float64(end - start)
		}

		freqRatio := float64(i) / 15.0

		if freqRatio < 0.3 {
			bandAmplitude *= 1.2 - (freqRatio * 0.5)
		} else if freqRatio < 0.7 {
			bandAmplitude *= 1.0
		} else {
			bandAmplitude *= 0.8 - (freqRatio * 0.3)
		}

		bands[i] = math.Min(bandAmplitude*3, 1.0)
		if bands[i] < 0 {
			bands[i] = 0
		}
	}

	return bands
}
