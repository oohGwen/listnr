package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rivo/tview"
)

// Create app instance
func NewApp() *App {
	return &App{
		app:          tview.NewApplication(),
		focusMode:    "explorer",
		selectedSong: 0,
		selectedDir:  0,
		player:       &Player{},
	}
}

// starts the application
func (a *App) Run() error {
	return a.app.SetRoot(a.layout, true).EnableMouse(true).Run()
}

// loads the application configuration
func (a *App) LoadConfig() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Failed to get current user:", err)
	}

	configPath := filepath.Join(usr.HomeDir, ".config", "listenr.json")

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &Config{
			MusicRoutes: []string{filepath.Join(usr.HomeDir, "Music")},
			Volume:      0.5,
			LastPath:    "",
		}

		// Create .config directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			log.Fatal("Failed to create config directory:", err)
		}

		// Write default config
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			log.Fatal("Failed to marshal default config:", err)
		}

		if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
			log.Fatal("Failed to write default config:", err)
		}

		a.config = defaultConfig
		return
	}

	// Load existing config
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal("Failed to read config:", err)
	}

	a.config = &Config{}
	if err := json.Unmarshal(data, a.config); err != nil {
		log.Fatal("Failed to unmarshal config:", err)
	}
}

// scans all configured music directories
func (a *App) ScanDirectories() {
	a.directories = make([]Directory, 0)

	for _, route := range a.config.MusicRoutes {
		if dir := a.scanDirectory(route); dir != nil {
			a.directories = append(a.directories, *dir)
		}
	}

	if len(a.directories) > 0 {
		a.currentDir = &a.directories[0]
	}
}

// recursively scans a directory for music files
func (a *App) scanDirectory(path string) *Directory {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil
	}

	dir := &Directory{
		Path:  path,
		Name:  filepath.Base(path),
		Songs: make([]Song, 0),
		Dirs:  make([]Directory, 0),
	}

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return dir
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			if subDir := a.scanDirectory(fullPath); subDir != nil {
				dir.Dirs = append(dir.Dirs, *subDir)
			}
		} else {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if supportedExts[ext] {
				song := Song{
					Path: fullPath,
					Name: strings.TrimSuffix(entry.Name(), ext),
				}
				dir.Songs = append(dir.Songs, song)
			}
		}
	}

	// Sort directories and songs
	sort.Slice(dir.Dirs, func(i, j int) bool {
		return dir.Dirs[i].Name < dir.Dirs[j].Name
	})
	sort.Slice(dir.Songs, func(i, j int) bool {
		return dir.Songs[i].Name < dir.Songs[j].Name
	})

	return dir
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
