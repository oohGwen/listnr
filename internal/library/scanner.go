package library

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Scanner struct {
	supportedExts map[string]bool
}

func NewScanner() *Scanner {
	return &Scanner{
		supportedExts: map[string]bool{
			".mp3":  true,
			".wav":  true,
			".flac": true,
			".ogg":  true,
			".m4a":  true,
		},
	}
}

func (s *Scanner) ScanDirectory(path string) (*Directory, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, err
	}

	dir := &Directory{
		Path:  path,
		Name:  filepath.Base(path),
		Songs: make([]*Song, 0),
		Dirs:  make([]*Directory, 0),
	}

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return dir, nil
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			if subDir, err := s.ScanDirectory(fullPath); err == nil && subDir != nil {
				// Only add directory if it has songs or subdirectories
				if len(subDir.Songs) > 0 || len(subDir.Dirs) > 0 {
					dir.Dirs = append(dir.Dirs, subDir)
				}
			}
		} else {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if s.IsSupported(ext) {
				song := &Song{
					Path: fullPath,
					Name: strings.TrimSuffix(entry.Name(), ext),
				}
				dir.Songs = append(dir.Songs, song)
			}
		}
	}

	// Sort directories and songs alphabetically
	sort.Slice(dir.Dirs, func(i, j int) bool {
		return dir.Dirs[i].Name < dir.Dirs[j].Name
	})
	sort.Slice(dir.Songs, func(i, j int) bool {
		return dir.Songs[i].Name < dir.Songs[j].Name
	})

	return dir, nil
}

func (s *Scanner) IsSupported(ext string) bool {
	return s.supportedExts[strings.ToLower(ext)]
}

func (s *Scanner) ScanRecursive(paths []string) ([]*Directory, error) {
	var directories []*Directory

	for _, path := range paths {
		if dir, err := s.ScanDirectory(path); err == nil && dir != nil {
			if len(dir.Songs) > 0 || len(dir.Dirs) > 0 {
				directories = append(directories, dir)
			}
		}
	}

	return directories, nil
}
