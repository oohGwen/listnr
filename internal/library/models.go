package library

import "time"

type Song struct {
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Artist   string        `json:"artist,omitempty"`
	Album    string        `json:"album,omitempty"`
}

type Directory struct {
	Path  string       `json:"path"`
	Name  string       `json:"name"`
	Songs []*Song      `json:"songs"`
	Dirs  []*Directory `json:"dirs"`
}

func (d *Directory) GetAllSongs() []*Song {
	var songs []*Song
	songs = append(songs, d.Songs...)

	for _, subDir := range d.Dirs {
		songs = append(songs, subDir.GetAllSongs()...)
	}

	return songs
}

func (d *Directory) FindSong(path string) *Song {
	for _, song := range d.Songs {
		if song.Path == path {
			return song
		}
	}

	for _, subDir := range d.Dirs {
		if song := subDir.FindSong(path); song != nil {
			return song
		}
	}

	return nil
}
