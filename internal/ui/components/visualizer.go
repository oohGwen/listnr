package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Visualizer struct {
	TextView  *tview.TextView
	bars      []float64
	isPlaying bool
	amplitude float64
}

func NewVisualizer() *Visualizer {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetScrollable(false).
		SetWrap(false).
		SetBorder(true).
		SetBorderColor(tcell.ColorGray).
		SetTitle(" Visualizer ")

	textView.SetDisabled(true)
	textView.SetTextAlign(tview.AlignCenter)

	visualizer := &Visualizer{
		TextView:  textView,
		bars:      make([]float64, 16), // 16 frequency bands
		isPlaying: false,
		amplitude: 0.0,
	}

	return visualizer
}

// Update with audio data
func (v *Visualizer) UpdateAudioData(frequencyBands []float64, amplitude float64, isPlaying bool) {
	v.isPlaying = isPlaying
	v.amplitude = amplitude

	// If no data or playback stopped, apply gradual decay
	if !isPlaying || len(frequencyBands) == 0 {
		for i := range v.bars {
			v.bars[i] *= 0.85 // Faster decay
			if v.bars[i] < 0.02 {
				v.bars[i] = 0
			}
		}
	} else {
		// Use frequency data
		bandsToUse := len(frequencyBands)
		if bandsToUse > len(v.bars) {
			bandsToUse = len(v.bars)
		}

		for i := 0; i < bandsToUse; i++ {
			target := frequencyBands[i]

			// Smooth transitions
			if target > v.bars[i] {
				v.bars[i] += (target - v.bars[i]) * 0.4
			} else {
				v.bars[i] += (target - v.bars[i]) * 0.8
			}

			// Clamp values
			if v.bars[i] < 0 {
				v.bars[i] = 0
			}
			if v.bars[i] > 1 {
				v.bars[i] = 1
			}
		}
	}

	v.render()
}

func (v *Visualizer) render() {
	output := v.renderCompactBars()
	v.TextView.SetText(output)
}

func (v *Visualizer) renderCompactBars() string {
	_, _, width, height := v.TextView.GetInnerRect()

	if width < 16 || height < 2 {
		if v.isPlaying {
			return "[green]♪ Playing...[-]"
		} else {
			return "[dim]♪ Paused[-]"
		}
	}

	if !v.isPlaying {
		return "[dim]♪ ----[-]"
	}

	var line1, line2 strings.Builder
	charsPerBar := width / len(v.bars)
	if charsPerBar < 1 {
		charsPerBar = 1
	}

	for _, intensity := range v.bars {
		color := v.getBarColor(intensity)

		// Split intensity between bottom (0-0.5) and top (0.5-1.0)
		bottomLevel := intensity
		if bottomLevel > 0.5 {
			bottomLevel = 0.5
		}
		topLevel := intensity - bottomLevel
		if topLevel < 0 {
			topLevel = 0
		}

		// bottom row shading
		line2.WriteString(strings.Repeat(v.getCharForLevel(bottomLevel*2, color, true), charsPerBar))

		// top row shading
		line1.WriteString(strings.Repeat(v.getCharForLevel(topLevel*2, color, false), charsPerBar))
	}

	return line1.String() + "\n" + line2.String()
}

func (v *Visualizer) getCharForLevel(level float64, color string, base bool) string {
	if level <= 0.05 {
		if base {
			return "[dim]_[-]"
		} else {
			return ""
		}
	} else if level < 0.25 {
		return fmt.Sprintf("[%s]▁[-]", color)
	} else if level < 0.5 {
		return fmt.Sprintf("[%s]▃[-]", color)
	} else if level < 0.75 {
		return fmt.Sprintf("[%s]▅[-]", color)
	} else {
		return fmt.Sprintf("[%s]█[-]", color)
	}
}

func (v *Visualizer) getBarColor(intensity float64) string {
	if intensity < 0.2 {
		return "blue"
	} else if intensity < 0.4 {
		return "cyan"
	} else if intensity < 0.6 {
		return "green"
	} else if intensity < 0.8 {
		return "yellow"
	} else {
		return "red"
	}
}

// Cleanup method
func (v *Visualizer) Stop() {
	// Nothing to clean up for now
}
