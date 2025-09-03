package library

type Library struct {
	Directories []*Directory
	scanner     *Scanner
}

func NewLibrary() *Library {
	return &Library{
		Directories: make([]*Directory, 0),
		scanner:     NewScanner(),
	}
}

func (l *Library) Scan(paths []string) error {
	dirs, err := l.scanner.ScanRecursive(paths)
	if err != nil {
		return err
	}

	l.Directories = dirs
	return nil
}

func (l *Library) FindSong(path string) (*Song, error) {
	for _, dir := range l.Directories {
		if song := dir.FindSong(path); song != nil {
			return song, nil
		}
	}
	return nil, nil
}

func (l *Library) GetAllSongs() []*Song {
	var songs []*Song
	for _, dir := range l.Directories {
		songs = append(songs, dir.GetAllSongs()...)
	}
	return songs
}

func (l *Library) GetDirectories() []*Directory {
	return l.Directories
}
