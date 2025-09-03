package components

import (
	"strings"

	"listnr/internal/library"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Sidebar struct {
	List              *tview.List
	directories       []*library.Directory
	selectionCallback func(*library.Directory)
}

func NewSidebar() *Sidebar {
	list := tview.NewList()
	list.ShowSecondaryText(false).
		SetHighlightFullLine(true).
		SetBorder(true).
		SetTitle(" Listnr ")

	return &Sidebar{
		List: list,
	}
}

func (s *Sidebar) SetDirectories(directories []*library.Directory) {
	s.directories = directories
	s.populateList()
}

func (s *Sidebar) SetSelectionCallback(callback func(*library.Directory)) {
	s.selectionCallback = callback
}

func (s *Sidebar) populateList() {
	s.List.Clear()

	var addDirToList func(*library.Directory, int)
	addDirToList = func(dir *library.Directory, level int) {
		indent := strings.Repeat("  ", level)
		displayName := indent + "📁 " + dir.Name

		s.List.AddItem(displayName, "", 0, func() {
			if s.selectionCallback != nil {
				s.selectionCallback(dir)
			}
		})

		for _, subDir := range dir.Dirs {
			addDirToList(subDir, level+1)
		}
	}

	for _, dir := range s.directories {
		addDirToList(dir, 0)
	}
}

func (s *Sidebar) SetFocused(focused bool) {
	if focused {
		s.List.SetBorderColor(tcell.ColorWhite)
	} else {
		s.List.SetBorderColor(tcell.ColorGray)
	}
}
