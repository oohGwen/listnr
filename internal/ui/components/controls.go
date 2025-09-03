package components

import (
	"fmt"
	"strings"
	"time"

	"listnr/internal/library"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Controls struct {
	TextView        *tview.TextView
	currentSong     *library.Song
	isPlaying       bool
	volume          float64
	position        time.Duration
	duration        time.Duration
	autoplayEnabled bool
	repeatMode      bool
}

func NewControls() *Controls {
	textView := tview.NewTextView()
	textView.SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetScrollable(false).
		SetWrap(false).
		SetBorder(true).
		SetBorderColor(tcell.ColorGray)
	textView.SetDisabled(true)

	controls := &Controls{
		TextView: textView,
		volume:   0.5,
	}

	controls.update()
	return controls
}

func (c *Controls) SetCurrentSong(song *library.Song) {
	c.currentSong = song
	c.update()
}

func (c *Controls) SetPlaying(playing bool) {
	c.isPlaying = playing
	c.update()
}

func (c *Controls) SetVolume(volume float64) {
	c.volume = volume
	c.update()
}

func (c *Controls) SetAutoplay(enabled bool) {
	c.autoplayEnabled = enabled
	c.update()
}

func (c *Controls) SetRepeatMode(enabled bool) {
	c.repeatMode = enabled
	c.update()
}

func (c *Controls) UpdateProgress(position, duration time.Duration) {
	c.position = position
	c.duration = duration
	c.update()
}

func (c *Controls) update() {
	text := c.getControlsText()
	c.TextView.SetText(text)

	if c.currentSong != nil {
		title := fmt.Sprintf(" %s ", c.currentSong.Name)
		c.TextView.SetTitle(title)
	} else {
		c.TextView.SetTitle(" <No Playing> ")
	}
}

func (c *Controls) getControlsText() string {
	line1 := c.getProgressLine()
	line2 := c.getControlsLine()
	return line1 + "\n" + line2
}

func (c *Controls) getProgressLine() string {
	var currentTime, totalTime string
	var progress float64

	if c.duration == 0 {
		currentTime = "00:00"
		totalTime = "--:--"
		progress = 0.0
	} else {
		currentSeconds := int(c.position.Seconds())
		totalSeconds := int(c.duration.Seconds())

		currentTime = fmt.Sprintf("%02d:%02d", currentSeconds/60, currentSeconds%60)
		totalTime = fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
		progress = float64(currentSeconds) / float64(totalSeconds)

		if progress > 1.0 {
			progress = 1.0
		}
	}

	// Calculate bar width dynamically
	_, totalWidth, _, _ := c.TextView.GetInnerRect()
	timeWidth := len(currentTime) + len(totalTime) + 2
	barWidth := totalWidth - timeWidth

	if barWidth < 30 {
		barWidth = 30
	}

	// Build progress bar
	filledWidth := int(progress * float64(barWidth))

	var bar strings.Builder
	bar.WriteString("[green]")

	for i := 0; i < barWidth; i++ {
		if i < filledWidth-1 {
			bar.WriteString("=")
		} else if i == filledWidth-1 && progress > 0 {
			bar.WriteString("●")
		} else {
			if i == filledWidth {
				bar.WriteString("[dim]")
			}
			bar.WriteString("-")
		}
	}
	bar.WriteString("[-]")

	return fmt.Sprintf("[cyan]%s[-] %s [cyan]%s[-]", currentTime, bar.String(), totalTime)
}

func (c *Controls) getControlsLine() string {
	// Player controls
	var playIcon string
	if c.isPlaying {
		playIcon = "⏸"
	} else {
		playIcon = "▶"
	}

	var repeatIcon, autoplayIcon string
	if c.repeatMode {
		repeatIcon = "[green][🔁 R][-]"
	} else {
		repeatIcon = "[red][🔁 R][-]"
	}

	if c.autoplayEnabled {
		autoplayIcon = "[green][⏭ N][-]"
	} else {
		autoplayIcon = "[red][⏭ N][-]"
	}

	controls := fmt.Sprintf(" %s %s   [⏮ Q] [⏪ A] [%s SPACE] [⏩ D] [⏭ E]  ",
		repeatIcon, autoplayIcon, playIcon)

	// Volume bar (10 segments)
	volumeSegments := int(c.volume * 10)

	var volBar strings.Builder
	volBar.WriteString("[♪ ")
	for i := 0; i < 10; i++ {
		if i < volumeSegments {
			volBar.WriteString("[magenta]■[-]")
		} else {
			volBar.WriteString("[dim]□[-]")
		}
	}
	volBar.WriteString(fmt.Sprintf(" %d%% W/S]", int(c.volume*100)))

	volumeStr := volBar.String()

	// Calculate spacing
	controlsLen := len(controls) - 6 // Account for color tags
	volumeLen := 20
	_, totalWidth, _, _ := c.TextView.GetInnerRect()

	spacing := totalWidth - controlsLen - volumeLen
	if spacing < 1 {
		spacing = 1
	}

	return fmt.Sprintf("%s%s%s", controls, strings.Repeat(" ", spacing), volumeStr)
}
