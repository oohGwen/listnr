package components

import (
	"fmt"

	"listnr/internal/library"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SongList struct {
	List              *tview.List
	directory         *library.Directory
	selectionCallback func(*library.Song, int)
}

func NewSongList() *SongList {
	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetBorder(true).
		SetTitle(" Songs ")

	return &SongList{
		List: list,
	}
}

func (sl *SongList) SetDirectory(directory *library.Directory) {
	sl.directory = directory
	sl.populateList()
}

func (sl *SongList) SetSelectionCallback(callback func(*library.Song, int)) {
	sl.selectionCallback = callback
}

func (sl *SongList) populateList() {
	sl.List.Clear()

	if sl.directory == nil {
		return
	}

	for i, song := range sl.directory.Songs {
		displayName := "🎵 " + song.Name
		// Capture variables for closure
		currentSong := song
		currentIndex := i

		sl.List.AddItem(displayName, "", 0, func() {
			if sl.selectionCallback != nil {
				sl.selectionCallback(currentSong, currentIndex)
			}
		})
	}

	// Update title with current directory
	sl.List.SetTitle(fmt.Sprintf(" Songs - %s ", sl.directory.Name))
}

func (sl *SongList) SetCurrentItem(index int) {
	if index >= 0 && index < sl.List.GetItemCount() {
		sl.List.SetCurrentItem(index)
	}
}

func (sl *SongList) SetFocused(focused bool) {
	if focused {
		sl.List.SetBorderColor(tcell.ColorWhite)
	} else {
		sl.List.SetBorderColor(tcell.ColorGray)
	}
}
