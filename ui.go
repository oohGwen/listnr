package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// initializes all UI components and layout:
// [S] [L]   -> sidebar / song list
// [  P  ]   -> player controls
func (a *App) SetupUI() {
	// Sidebar (directory tree)
	a.sidebar = tview.NewList()
	a.sidebar.ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetBorder(true).
		SetTitle(" Listnr ")

	// Song list
	a.songList = tview.NewList()
	a.songList.
		ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetBorder(true).
		SetTitle(" Songs ")

	// Controls bar - disable scrolling and set as non-focusable
	a.controls = tview.NewTextView()
	a.controls.
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetScrollable(false).
		SetWrap(false).
		SetBorder(true)

	// Populate sidebar
	a.populateSidebar()

	// Populate song list
	a.populateSongList()

	// Layout setup
	topLayout := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.sidebar, 0, 1, true).
		AddItem(a.songList, 0, 3, false)

	a.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topLayout, 0, 1, true).
		AddItem(a.controls, 4, 0, false)

	// Set initial focus
	a.app.SetFocus(a.sidebar)
	a.UpdateControlsDisplay()
}

// fills the sidebar with directory tree
func (a *App) populateSidebar() {
	a.sidebar.Clear()

	var addDirToList func(*Directory, int)
	addDirToList = func(dir *Directory, level int) {
		indent := strings.Repeat("  ", level)
		displayName := indent + "📁 " + dir.Name

		a.sidebar.AddItem(displayName, "", 0, func() {
			a.currentDir = dir
			a.selectedSong = 0
			a.populateSongList()
		})

		for _, subDir := range dir.Dirs {
			addDirToList(&subDir, level+1)
		}
	}

	for _, dir := range a.directories {
		addDirToList(&dir, 0)
	}
}

// fills the song list with songs from current directory
func (a *App) populateSongList() {
	a.songList.Clear()

	if a.currentDir == nil {
		return
	}

	for i, song := range a.currentDir.Songs {
		displayName := "🎵 " + song.Name
		// Capture the song reference correctly
		currentSong := song
		currentIndex := i

		a.songList.AddItem(displayName, "", 0, func() {
			a.selectedSong = currentIndex
			a.PlaySong(&currentSong)
		})
	}

	// Update title with current directory
	a.songList.SetTitle(fmt.Sprintf(" Songs - %s ", a.currentDir.Name))
}

// returns exactly 2 lines for the player controls
func (a *App) GetControlsText() string {
	// Line 1: Time + Progress Bar + Time
	line1 := a.getProgressLine()

	// Line 2: Centered controls + Volume bar
	line2 := a.getControlsLine()

	return line1 + "\n" + line2
}

// creates the first line: [TIME] [========●--------] [TIME]
func (a *App) getProgressLine() string {
	var currentTime, totalTime string
	var progress float64

	if a.player.streamer == nil || a.player.format.SampleRate == 0 {
		currentTime = "00:00"
		totalTime = "--:--"
		progress = 0.0
	} else {
		// Frames
		currentPos := a.player.streamer.Position()
		totalFrames := a.player.streamer.Len()

		// Convert to seconds
		currentSeconds := int(currentPos) / int(a.player.format.SampleRate)
		totalSeconds := int(totalFrames) / int(a.player.format.SampleRate)

		currentTime = fmt.Sprintf("%02d:%02d", currentSeconds/60, currentSeconds%60)
		if totalSeconds > 0 {
			totalTime = fmt.Sprintf("%02d:%02d", totalSeconds/60, totalSeconds%60)
			progress = float64(currentSeconds) / float64(totalSeconds)
		} else {
			totalTime = "--:--"
			progress = 0.0
		}

		if progress > 1.0 {
			progress = 1.0
		}
	}

	// Dynamically calculate bar width
	_, totalWidth, _, _ := a.controls.GetInnerRect()
	timeWidth := len(currentTime) + len(totalTime) + 2 // space between them
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

// creates the second line: centered controls + volume bar
func (a *App) getControlsLine() string {
	// Player controls
	var playIcon string
	if a.player.isPlaying {
		playIcon = "⏸"
	} else {
		playIcon = "▶"
	}

	controls := fmt.Sprintf("[⏮ A] [%s SPACE] [⏭ D]", playIcon)

	// Volume bar (10 segments for 0-100%, each segment = 10%)
	volumeLevel := a.config.Volume
	volumeSegments := int(volumeLevel * 10) // 0-10 segments

	var volBar strings.Builder
	volBar.WriteString("[♪ ")
	for i := 0; i < 10; i++ {
		if i < volumeSegments {
			volBar.WriteString("[magenta]■[-]")
		} else {
			volBar.WriteString("[dim]□[-]")
		}
	}
	volBar.WriteString(fmt.Sprintf(" %d%% ↑↓]", int(volumeLevel*100)))

	volumeStr := volBar.String()

	// Calculate spacing to center controls and right-align volume
	controlsLen := len(controls) - 6
	volumeLen := 20
	_, totalWidth, _, _ := a.controls.GetInnerRect()

	spacing := totalWidth - controlsLen - volumeLen
	if spacing < 1 {
		spacing = 1
	}

	return fmt.Sprintf("%s%s%s", controls, strings.Repeat(" ", spacing), volumeStr)
}

// updates the controls display
func (a *App) UpdateControlsDisplay() {
	a.controls.SetText(a.GetControlsText())

	if a.player.currentSong != nil {
		display := fmt.Sprintf(" %s ", a.player.currentSong.Name)
		a.controls.SetTitle(display)
	} else {
		a.controls.SetTitle(" <No Playing> ")
	}

	// Update focus indicator - only show yellow border when in controls mode
	if a.focusMode == "controls" {
		a.controls.SetBorderColor(tcell.ColorYellow)
		a.sidebar.SetBorderColor(tcell.ColorDarkGray)
		a.songList.SetBorderColor(tcell.ColorDarkGray)
	} else {
		a.controls.SetBorderColor(tcell.ColorWhite)
		// Reset sidebar and songlist colors to default
		a.sidebar.SetBorderColor(tcell.ColorWhite)
		a.songList.SetBorderColor(tcell.ColorWhite)
	}
}
