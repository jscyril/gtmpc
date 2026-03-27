# Terminal Music Player - System Design Document

A comprehensive Go-based terminal music player showcasing advanced Go concepts including concurrency, pointers, interfaces, and more.

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture Design](#architecture-design)
3. [Core Data Structures](#core-data-structures)
4. [Interfaces Design](#interfaces-design)
5. [Concurrency Model](#concurrency-model)
6. [Pointer Usage Patterns](#pointer-usage-patterns)
7. [JSON Handling & Unit Tests](#json-handling--unit-tests)
8. [Error Handling Strategy](#error-handling-strategy)
9. [Maps, Arrays & Slices Usage](#maps-arrays--slices-usage)
10. [Project Structure](#project-structure)
11. [Implementation Roadmap](#implementation-roadmap)
12. [Client-Server Architecture](#client-server-architecture)
13. [HTTP Server & Routing Design](#http-server--routing-design)
14. [Security Architecture (bcrypt)](#security-architecture-bcrypt)
15. [Server Concurrency Model](#server-concurrency-model)
16. [API Reference](#api-reference)
17. [PostgreSQL Database Layer](#postgresql-database-layer)

---

## Project Overview

### Goal
Build a feature-rich terminal-based music player that supports:
- Playing audio files (MP3, WAV, FLAC)
- Playlist management (create, save, load)
- Music library scanning and indexing
- Real-time playback controls (play, pause, stop, next, previous)
- Volume control and seek functionality
- Search and filter capabilities
- Keyboard-driven TUI (Terminal User Interface)

### Tech Stack
| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Audio Playback | [beep](https://github.com/faiface/beep) or [oto](https://github.com/hajimehoshi/oto) |
| TUI Framework | [bubbletea](https://github.com/charmbracelet/bubbletea) or [tview](https://github.com/rivo/tview) |
| Configuration | JSON files |
| Metadata | [tag](https://github.com/dhowden/tag) for ID3 tags |

---

## Architecture Design

```
┌─────────────────────────────────────────────────────────────────┐
│                        Terminal UI Layer                        │
│                    (bubbletea / tview)                          │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Controller Layer                            │
│         (Handles user input, coordinates components)             │
└──────────┬──────────────────┬──────────────────┬────────────────┘
           │                  │                  │
           ▼                  ▼                  ▼
┌──────────────────┐ ┌─────────────────┐ ┌─────────────────────────┐
│  Audio Engine    │ │ Library Manager │ │   Playlist Manager      │
│  (Goroutines)    │ │  (Goroutines)   │ │                         │
│                  │ │                 │ │                         │
│ - Playback       │ │ - Scan files    │ │ - CRUD operations       │
│ - Volume control │ │ - Index tracks  │ │ - JSON serialization    │
│ - Seek           │ │ - Search        │ │ - Load/Save             │
└────────┬─────────┘ └────────┬────────┘ └────────┬────────────────┘
         │                    │                   │
         └────────────────────┴───────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Storage Layer                             │
│              (File System, JSON Config, Cache)                   │
└─────────────────────────────────────────────────────────────────┘
```

### Component Communication Flow

```
┌─────────────┐    Commands (chan)     ┌─────────────────┐
│   UI Layer  │ ─────────────────────► │  Audio Engine   │
│             │ ◄───────────────────── │   (Goroutine)   │
└─────────────┘    State Updates       └─────────────────┘
       │                                       │
       │ Events (chan)                         │ Playback Events
       ▼                                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Event Bus (Channels)                         │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Data Structures

### Structs

```go
// Track represents a single audio file with metadata
type Track struct {
    ID        string        `json:"id"`
    Title     string        `json:"title"`
    Artist    string        `json:"artist"`
    Album     string        `json:"album"`
    Duration  time.Duration `json:"duration"`
    FilePath  string        `json:"file_path"`
    Genre     string        `json:"genre"`
    Year      int           `json:"year"`
    TrackNum  int           `json:"track_number"`
    CoverArt  []byte        `json:"-"` // Excluded from JSON
    CreatedAt time.Time     `json:"created_at"`
}

// Playlist represents a collection of tracks
type Playlist struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Tracks      []Track   `json:"tracks"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// Library represents the entire music collection
type Library struct {
    Tracks       map[string]*Track    `json:"tracks"`        // Map by ID for O(1) lookup
    Playlists    map[string]*Playlist `json:"playlists"`     // Map by ID
    Artists      map[string][]string  `json:"artists"`       // Artist -> []TrackID
    Albums       map[string][]string  `json:"albums"`        // Album -> []TrackID
    ScanPaths    []string             `json:"scan_paths"`
    LastScanned  time.Time            `json:"last_scanned"`
    TotalTracks  int                  `json:"total_tracks"`
}

// PlaybackState represents the current player state
type PlaybackState struct {
    CurrentTrack   *Track        `json:"current_track"`
    Status         PlayerStatus  `json:"status"` // Playing, Paused, Stopped
    Position       time.Duration `json:"position"`
    Volume         float64       `json:"volume"` // 0.0 to 1.0
    Repeat         RepeatMode    `json:"repeat"` // None, One, All
    Shuffle        bool          `json:"shuffle"`
    Queue          []*Track      `json:"queue"`
    QueueIndex     int           `json:"queue_index"`
}

// PlayerStatus represents playback status
type PlayerStatus int

const (
    StatusStopped PlayerStatus = iota
    StatusPlaying
    StatusPaused
)

// RepeatMode represents repeat options
type RepeatMode int

const (
    RepeatNone RepeatMode = iota
    RepeatOne
    RepeatAll
)

// Config holds application configuration
type Config struct {
    MusicDirectories []string `json:"music_directories"`
    DefaultVolume    float64  `json:"default_volume"`
    Theme            string   `json:"theme"`
    KeyBindings      KeyMap   `json:"key_bindings"`
    EnableCache      bool     `json:"enable_cache"`
    CachePath        string   `json:"cache_path"`
}

// KeyMap defines keyboard shortcuts
type KeyMap struct {
    PlayPause   string `json:"play_pause"`
    Stop        string `json:"stop"`
    Next        string `json:"next"`
    Previous    string `json:"previous"`
    VolumeUp    string `json:"volume_up"`
    VolumeDown  string `json:"volume_down"`
    SeekForward string `json:"seek_forward"`
    SeekBack    string `json:"seek_back"`
    Quit        string `json:"quit"`
}

// AudioCommand represents commands sent to the audio engine
type AudioCommand struct {
    Type    CommandType
    Payload interface{} // Can be *Track, float64 for volume, etc.
}

// CommandType enumerates audio commands
type CommandType int

const (
    CmdPlay CommandType = iota
    CmdPause
    CmdStop
    CmdSeek
    CmdVolume
    CmdNext
    CmdPrevious
)

// AudioEvent represents events emitted by the audio engine
type AudioEvent struct {
    Type    EventType
    Payload interface{}
}

// EventType enumerates audio events
type EventType int

const (
    EventTrackStarted EventType = iota
    EventTrackEnded
    EventPositionUpdate
    EventError
    EventStateChange
)
```

---

## Interfaces Design

Interfaces enable loose coupling and easy testing through dependency injection.

```go
// Player defines the core playback interface
type Player interface {
    Play(track *Track) error
    Pause() error
    Resume() error
    Stop() error
    Seek(position time.Duration) error
    SetVolume(level float64) error
    GetState() *PlaybackState
}

// LibraryScanner defines the interface for scanning music files
type LibraryScanner interface {
    Scan(paths []string) (<-chan *Track, <-chan error)
    ScanFile(path string) (*Track, error)
    SupportedFormats() []string
}

// PlaylistRepository defines CRUD operations for playlists
type PlaylistRepository interface {
    Create(playlist *Playlist) error
    GetByID(id string) (*Playlist, error)
    GetAll() ([]*Playlist, error)
    Update(playlist *Playlist) error
    Delete(id string) error
    AddTrack(playlistID string, track *Track) error
    RemoveTrack(playlistID string, trackID string) error
}

// TrackRepository defines CRUD operations for tracks
type TrackRepository interface {
    Add(track *Track) error
    GetByID(id string) (*Track, error)
    GetAll() ([]*Track, error)
    Search(query string) ([]*Track, error)
    GetByArtist(artist string) ([]*Track, error)
    GetByAlbum(album string) ([]*Track, error)
    Delete(id string) error
}

// ConfigManager handles application configuration
type ConfigManager interface {
    Load() (*Config, error)
    Save(config *Config) error
    GetDefault() *Config
}

// MetadataReader extracts metadata from audio files
type MetadataReader interface {
    Read(filePath string) (*Track, error)
    ReadCoverArt(filePath string) ([]byte, error)
}

// EventBus handles event distribution
type EventBus interface {
    Subscribe(eventType EventType) <-chan AudioEvent
    Publish(event AudioEvent)
    Unsubscribe(ch <-chan AudioEvent)
}

// UI defines the terminal user interface contract
type UI interface {
    Init() error
    Run() error
    Update(state *PlaybackState)
    ShowError(err error)
    Close() error
}
```

### Interface Implementation Example

```go
// Ensure AudioEngine implements Player interface at compile time
var _ Player = (*AudioEngine)(nil)

type AudioEngine struct {
    state      *PlaybackState
    commands   chan AudioCommand
    events     chan AudioEvent
    mu         sync.RWMutex
    streamer   beep.StreamSeekCloser
    ctrl       *beep.Ctrl
    volume     *effects.Volume
}

func NewAudioEngine() *AudioEngine {
    return &AudioEngine{
        state:    &PlaybackState{Status: StatusStopped, Volume: 0.5},
        commands: make(chan AudioCommand, 10),
        events:   make(chan AudioEvent, 10),
    }
}
```

---

## Concurrency Model

### Goroutines Architecture

```
Main Goroutine
     │
     ├──► Audio Engine Goroutine (long-running)
     │         │
     │         ├──► Playback Worker
     │         └──► Position Tracker (ticker-based)
     │
     ├──► Library Scanner Goroutine (on-demand)
     │         │
     │         └──► File Scanner Workers (worker pool)
     │
     ├──► Event Dispatcher Goroutine
     │
     └──► UI Event Loop (bubbletea handles this)
```

### Channel Design

```go
// Command channel - buffered to prevent blocking UI
commands := make(chan AudioCommand, 10)

// Event channel - buffered for smooth event flow
events := make(chan AudioEvent, 20)

// Progress updates - ticker-based position updates
positionUpdates := make(chan time.Duration, 1)

// Scan results - unbuffered for controlled flow
scanResults := make(chan *Track)
scanErrors := make(chan error)

// Shutdown signal - signals goroutines to stop
shutdown := make(chan struct{})
```

### Audio Engine Implementation

```go
// AudioEngine manages playback in a separate goroutine
func (e *AudioEngine) Start(ctx context.Context) {
    go e.run(ctx)
    go e.trackPosition(ctx)
}

func (e *AudioEngine) run(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            e.cleanup()
            return
            
        case cmd := <-e.commands:
            switch cmd.Type {
            case CmdPlay:
                track := cmd.Payload.(*Track)
                if err := e.playTrack(track); err != nil {
                    e.events <- AudioEvent{Type: EventError, Payload: err}
                }
                
            case CmdPause:
                e.mu.Lock()
                if e.ctrl != nil {
                    e.ctrl.Paused = true
                    e.state.Status = StatusPaused
                }
                e.mu.Unlock()
                e.events <- AudioEvent{Type: EventStateChange, Payload: e.state}
                
            case CmdVolume:
                level := cmd.Payload.(float64)
                e.mu.Lock()
                if e.volume != nil {
                    e.volume.Volume = level
                    e.state.Volume = level
                }
                e.mu.Unlock()
                
            case CmdSeek:
                pos := cmd.Payload.(time.Duration)
                e.seekTo(pos)
            }
        }
    }
}

// trackPosition updates playback position periodically
func (e *AudioEngine) trackPosition(ctx context.Context) {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            e.mu.RLock()
            if e.state.Status == StatusPlaying && e.streamer != nil {
                pos := e.streamer.Position()
                e.state.Position = time.Duration(pos) * time.Second / time.Duration(e.sampleRate)
                e.events <- AudioEvent{
                    Type:    EventPositionUpdate,
                    Payload: e.state.Position,
                }
            }
            e.mu.RUnlock()
        }
    }
}
```

### Library Scanner with Worker Pool

```go
// Scanner scans directories concurrently using a worker pool
type Scanner struct {
    workers     int
    formats     []string
    metaReader  MetadataReader
}

func (s *Scanner) Scan(paths []string) (<-chan *Track, <-chan error) {
    tracks := make(chan *Track, 100)
    errors := make(chan error, 10)
    files := make(chan string, 100)
    
    var wg sync.WaitGroup
    
    // Start file discovery goroutine
    go func() {
        defer close(files)
        for _, path := range paths {
            filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
                if err != nil {
                    errors <- fmt.Errorf("walk error: %w", err)
                    return nil
                }
                if !d.IsDir() && s.isSupported(p) {
                    files <- p
                }
                return nil
            })
        }
    }()
    
    // Start worker pool
    for i := 0; i < s.workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for filePath := range files {
                track, err := s.metaReader.Read(filePath)
                if err != nil {
                    errors <- fmt.Errorf("read error for %s: %w", filePath, err)
                    continue
                }
                tracks <- track
            }
        }()
    }
    
    // Close channels when done
    go func() {
        wg.Wait()
        close(tracks)
        close(errors)
    }()
    
    return tracks, errors
}
```

### Channel Patterns Used

| Pattern | Use Case |
|---------|----------|
| **Fan-out** | File scanner distributes work to multiple workers |
| **Fan-in** | Multiple scanner workers send results to single channel |
| **Pipeline** | File discovery → Metadata reading → Library indexing |
| **Done channel** | Graceful shutdown with `ctx.Done()` |
| **Ticker** | Periodic position updates |
| **Select** | Multiplexing commands and context cancellation |

---

## Pointer Usage Patterns

### Call by Value vs Call by Reference

```go
// Call by Value - safe for small immutable data
// The function receives a COPY - changes don't affect original
func formatDuration(d time.Duration) string {
    minutes := int(d.Minutes())
    seconds := int(d.Seconds()) % 60
    return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// Usage: Original duration is unaffected
dur := 3 * time.Minute
formatted := formatDuration(dur) // dur is copied, not modified

// Call by Reference (Pointer) - for mutations and large structs
// The function receives a POINTER - changes affect original
func (t *Track) UpdateMetadata(artist, album string) {
    t.Artist = artist     // Modifies original Track
    t.Album = album       // Modifies original Track
    t.UpdatedAt = time.Now()
}

// Usage: Original track IS modified
track := &Track{Title: "Song"}
track.UpdateMetadata("New Artist", "New Album") // track is modified
```

### When to Use Pointers

```go
// 1. ALWAYS use pointers for method receivers that modify state
func (e *AudioEngine) SetVolume(level float64) error {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.state.Volume = level  // Modifies engine's state
    return nil
}

// 2. Use pointers for large structs to avoid copying overhead
func processLibrary(lib *Library) error {
    // Library might contain thousands of tracks
    // Passing by pointer avoids copying the entire map
    for id, track := range lib.Tracks {
        // Process each track...
    }
    return nil
}

// 3. Use value receivers for small, immutable operations
func (t Track) FullTitle() string {
    return fmt.Sprintf("%s - %s", t.Artist, t.Title)
}

// 4. Return pointers when creating new instances
func NewTrack(path string) (*Track, error) {
    track := &Track{
        ID:        generateID(),
        FilePath:  path,
        CreatedAt: time.Now(),
    }
    return track, nil
}

// 5. Use pointers in maps for efficient updates
type Library struct {
    Tracks map[string]*Track  // Pointer allows in-place updates
}

func (l *Library) UpdateTrack(id string, title string) {
    if track, ok := l.Tracks[id]; ok {
        track.Title = title  // Updates without reassigning map entry
    }
}
```

### Pointer Safety Patterns

```go
// Nil check before dereferencing
func (e *AudioEngine) GetCurrentTrack() *Track {
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    if e.state == nil || e.state.CurrentTrack == nil {
        return nil
    }
    
    // Return a copy to prevent external modification
    track := *e.state.CurrentTrack
    return &track
}

// Defensive copying for slice fields
func (p *Playlist) GetTracks() []Track {
    // Return a copy to prevent modification of internal slice
    tracks := make([]Track, len(p.Tracks))
    copy(tracks, p.Tracks)
    return tracks
}
```

---

## JSON Handling & Unit Tests

### JSON Marshalling/Unmarshalling

```go
// config.go - JSON configuration handling

// SaveConfig marshals and saves configuration to file
func SaveConfig(config *Config, path string) error {
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal config: %w", err)
    }
    
    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }
    
    return nil
}

// LoadConfig reads and unmarshals configuration from file
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return GetDefaultConfig(), nil
        }
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}

// GetDefaultConfig returns default configuration
func GetDefaultConfig() *Config {
    return &Config{
        MusicDirectories: []string{},
        DefaultVolume:    0.5,
        Theme:            "dark",
        EnableCache:      true,
        CachePath:        ".cache/musicplayer",
        KeyBindings: KeyMap{
            PlayPause:   "space",
            Stop:        "s",
            Next:        "n",
            Previous:    "p",
            VolumeUp:    "+",
            VolumeDown:  "-",
            SeekForward: "right",
            SeekBack:    "left",
            Quit:        "q",
        },
    }
}
```

### Playlist JSON Operations

```go
// playlist_repository.go

type JSONPlaylistRepository struct {
    basePath string
    mu       sync.RWMutex
}

func (r *JSONPlaylistRepository) Save(playlist *Playlist) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    playlist.UpdatedAt = time.Now()
    
    data, err := json.MarshalIndent(playlist, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal playlist: %w", err)
    }
    
    path := filepath.Join(r.basePath, playlist.ID+".json")
    return os.WriteFile(path, data, 0644)
}

func (r *JSONPlaylistRepository) Load(id string) (*Playlist, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    path := filepath.Join(r.basePath, id+".json")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read playlist file: %w", err)
    }
    
    var playlist Playlist
    if err := json.Unmarshal(data, &playlist); err != nil {
        return nil, fmt.Errorf("unmarshal playlist: %w", err)
    }
    
    return &playlist, nil
}
```

### Comprehensive Unit Tests

```go
// config_test.go

package config

import (
    "encoding/json"
    "os"
    "path/filepath"
    "testing"
    "time"
)

// TestConfigMarshal tests JSON marshalling of Config struct
func TestConfigMarshal(t *testing.T) {
    config := &Config{
        MusicDirectories: []string{"/home/user/Music", "/mnt/external/songs"},
        DefaultVolume:    0.75,
        Theme:            "dark",
        EnableCache:      true,
        CachePath:        ".cache/player",
        KeyBindings: KeyMap{
            PlayPause:  "space",
            Stop:       "s",
            Next:       "n",
            Previous:   "p",
            VolumeUp:   "+",
            VolumeDown: "-",
            Quit:       "q",
        },
    }
    
    data, err := json.Marshal(config)
    if err != nil {
        t.Fatalf("Failed to marshal config: %v", err)
    }
    
    // Verify JSON contains expected fields
    var result map[string]interface{}
    if err := json.Unmarshal(data, &result); err != nil {
        t.Fatalf("Failed to unmarshal result: %v", err)
    }
    
    if result["default_volume"].(float64) != 0.75 {
        t.Errorf("Expected volume 0.75, got %v", result["default_volume"])
    }
    
    if result["theme"].(string) != "dark" {
        t.Errorf("Expected theme 'dark', got %v", result["theme"])
    }
    
    dirs := result["music_directories"].([]interface{})
    if len(dirs) != 2 {
        t.Errorf("Expected 2 music directories, got %d", len(dirs))
    }
}

// TestConfigUnmarshal tests JSON unmarshalling of Config struct
func TestConfigUnmarshal(t *testing.T) {
    jsonData := `{
        "music_directories": ["/home/user/Music"],
        "default_volume": 0.8,
        "theme": "light",
        "enable_cache": false,
        "cache_path": "/tmp/cache",
        "key_bindings": {
            "play_pause": "p",
            "stop": "x",
            "next": ">",
            "previous": "<",
            "volume_up": "=",
            "volume_down": "-",
            "quit": "q"
        }
    }`
    
    var config Config
    if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
        t.Fatalf("Failed to unmarshal config: %v", err)
    }
    
    if config.DefaultVolume != 0.8 {
        t.Errorf("Expected volume 0.8, got %f", config.DefaultVolume)
    }
    
    if config.Theme != "light" {
        t.Errorf("Expected theme 'light', got %s", config.Theme)
    }
    
    if config.EnableCache != false {
        t.Errorf("Expected cache disabled, got enabled")
    }
    
    if len(config.MusicDirectories) != 1 {
        t.Errorf("Expected 1 directory, got %d", len(config.MusicDirectories))
    }
    
    if config.KeyBindings.PlayPause != "p" {
        t.Errorf("Expected play_pause 'p', got %s", config.KeyBindings.PlayPause)
    }
}

// TestConfigRoundTrip tests marshal -> unmarshal preserves data
func TestConfigRoundTrip(t *testing.T) {
    original := GetDefaultConfig()
    original.MusicDirectories = []string{"/test/path"}
    original.DefaultVolume = 0.65
    
    // Marshal
    data, err := json.Marshal(original)
    if err != nil {
        t.Fatalf("Marshal failed: %v", err)
    }
    
    // Unmarshal
    var restored Config
    if err := json.Unmarshal(data, &restored); err != nil {
        t.Fatalf("Unmarshal failed: %v", err)
    }
    
    // Compare
    if original.DefaultVolume != restored.DefaultVolume {
        t.Errorf("Volume mismatch: %f != %f", original.DefaultVolume, restored.DefaultVolume)
    }
    
    if original.Theme != restored.Theme {
        t.Errorf("Theme mismatch: %s != %s", original.Theme, restored.Theme)
    }
    
    if len(original.MusicDirectories) != len(restored.MusicDirectories) {
        t.Errorf("Directories count mismatch")
    }
}

// TestTrackMarshal tests Track JSON handling
func TestTrackMarshal(t *testing.T) {
    track := &Track{
        ID:       "track-001",
        Title:    "Test Song",
        Artist:   "Test Artist",
        Album:    "Test Album",
        Duration: 3*time.Minute + 45*time.Second,
        FilePath: "/music/test.mp3",
        Genre:    "Rock",
        Year:     2024,
        TrackNum: 5,
        CoverArt: []byte("fake-image-data"), // Should be excluded
    }
    
    data, err := json.Marshal(track)
    if err != nil {
        t.Fatalf("Marshal failed: %v", err)
    }
    
    // Verify CoverArt is excluded (json:"-")
    var result map[string]interface{}
    json.Unmarshal(data, &result)
    
    if _, exists := result["CoverArt"]; exists {
        t.Error("CoverArt should be excluded from JSON")
    }
    
    if result["artist"].(string) != "Test Artist" {
        t.Errorf("Artist mismatch")
    }
}

// TestPlaylistMarshalUnmarshal tests Playlist JSON operations
func TestPlaylistMarshalUnmarshal(t *testing.T) {
    playlist := &Playlist{
        ID:          "playlist-001",
        Name:        "My Favorites",
        Description: "Best songs collection",
        Tracks: []Track{
            {ID: "t1", Title: "Song 1", Artist: "Artist 1"},
            {ID: "t2", Title: "Song 2", Artist: "Artist 2"},
        },
        CreatedAt: time.Now().UTC().Truncate(time.Second),
        UpdatedAt: time.Now().UTC().Truncate(time.Second),
    }
    
    data, err := json.MarshalIndent(playlist, "", "  ")
    if err != nil {
        t.Fatalf("Marshal failed: %v", err)
    }
    
    var restored Playlist
    if err := json.Unmarshal(data, &restored); err != nil {
        t.Fatalf("Unmarshal failed: %v", err)
    }
    
    if restored.Name != playlist.Name {
        t.Errorf("Name mismatch: %s != %s", restored.Name, playlist.Name)
    }
    
    if len(restored.Tracks) != 2 {
        t.Errorf("Expected 2 tracks, got %d", len(restored.Tracks))
    }
    
    if restored.Tracks[0].Title != "Song 1" {
        t.Errorf("First track title mismatch")
    }
}

// TestSaveLoadConfig tests file operations
func TestSaveLoadConfig(t *testing.T) {
    tempDir := t.TempDir()
    configPath := filepath.Join(tempDir, "config.json")
    
    original := &Config{
        MusicDirectories: []string{"/test/music"},
        DefaultVolume:    0.9,
        Theme:            "custom",
        EnableCache:      true,
    }
    
    // Save
    if err := SaveConfig(original, configPath); err != nil {
        t.Fatalf("Save failed: %v", err)
    }
    
    // Verify file exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        t.Fatal("Config file was not created")
    }
    
    // Load
    loaded, err := LoadConfig(configPath)
    if err != nil {
        t.Fatalf("Load failed: %v", err)
    }
    
    if loaded.DefaultVolume != original.DefaultVolume {
        t.Errorf("Volume mismatch after load")
    }
    
    if loaded.Theme != original.Theme {
        t.Errorf("Theme mismatch after load")
    }
}

// TestLoadConfigNotExists tests loading non-existent config
func TestLoadConfigNotExists(t *testing.T) {
    config, err := LoadConfig("/non/existent/path.json")
    if err != nil {
        t.Fatalf("Should return default on missing file: %v", err)
    }
    
    expected := GetDefaultConfig()
    if config.DefaultVolume != expected.DefaultVolume {
        t.Error("Should return default config values")
    }
}

// TestInvalidJSON tests error handling for malformed JSON
func TestInvalidJSON(t *testing.T) {
    invalidJSON := `{"music_directories": [1, 2, 3], "volume":}`
    
    var config Config
    err := json.Unmarshal([]byte(invalidJSON), &config)
    
    if err == nil {
        t.Error("Expected error for invalid JSON")
    }
}

// Table-driven tests for edge cases
func TestConfigVolumeValidation(t *testing.T) {
    tests := []struct {
        name     string
        json     string
        expected float64
    }{
        {
            name:     "normal volume",
            json:     `{"default_volume": 0.5}`,
            expected: 0.5,
        },
        {
            name:     "zero volume",
            json:     `{"default_volume": 0}`,
            expected: 0,
        },
        {
            name:     "max volume",
            json:     `{"default_volume": 1.0}`,
            expected: 1.0,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var config Config
            if err := json.Unmarshal([]byte(tt.json), &config); err != nil {
                t.Fatalf("Unmarshal failed: %v", err)
            }
            
            if config.DefaultVolume != tt.expected {
                t.Errorf("Got %f, want %f", config.DefaultVolume, tt.expected)
            }
        })
    }
}
```

---

## Error Handling Strategy

### Custom Error Types

```go
// errors.go

package player

import (
    "errors"
    "fmt"
)

// Sentinel errors for common conditions
var (
    ErrTrackNotFound    = errors.New("track not found")
    ErrPlaylistNotFound = errors.New("playlist not found")
    ErrInvalidFormat    = errors.New("unsupported audio format")
    ErrPlaybackFailed   = errors.New("playback failed")
    ErrEmptyQueue       = errors.New("playback queue is empty")
    ErrInvalidVolume    = errors.New("volume must be between 0.0 and 1.0")
)

// PlayerError wraps errors with additional context
type PlayerError struct {
    Op      string // Operation that failed
    Track   string // Track ID if applicable
    Err     error  // Underlying error
}

func (e *PlayerError) Error() string {
    if e.Track != "" {
        return fmt.Sprintf("%s failed for track %s: %v", e.Op, e.Track, e.Err)
    }
    return fmt.Sprintf("%s failed: %v", e.Op, e.Err)
}

func (e *PlayerError) Unwrap() error {
    return e.Err
}

// NewPlayerError creates a new PlayerError
func NewPlayerError(op, track string, err error) *PlayerError {
    return &PlayerError{Op: op, Track: track, Err: err}
}

// ScanError represents an error during library scanning
type ScanError struct {
    Path string
    Err  error
}

func (e *ScanError) Error() string {
    return fmt.Sprintf("scan error at %s: %v", e.Path, e.Err)
}

func (e *ScanError) Unwrap() error {
    return e.Err
}

// ValidationError for input validation failures
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

### Error Handling Patterns

```go
// Wrapping errors with context
func (e *AudioEngine) loadTrack(track *Track) error {
    file, err := os.Open(track.FilePath)
    if err != nil {
        return fmt.Errorf("open audio file: %w", err)
    }
    defer file.Close()
    
    streamer, format, err := mp3.Decode(file)
    if err != nil {
        return NewPlayerError("decode", track.ID, err)
    }
    
    // ... continue loading
    return nil
}

// Checking error types
func handlePlaybackError(err error) {
    var playerErr *PlayerError
    if errors.As(err, &playerErr) {
        log.Printf("Player error during %s: %v", playerErr.Op, playerErr.Err)
    }
    
    if errors.Is(err, ErrInvalidFormat) {
        log.Println("Unsupported format - skipping track")
    }
}

// Graceful error recovery in goroutines
func (s *Scanner) worker(files <-chan string, results chan<- *Track, errs chan<- error) {
    for path := range files {
        track, err := s.processFile(path)
        if err != nil {
            // Send error but continue processing
            errs <- &ScanError{Path: path, Err: err}
            continue
        }
        results <- track
    }
}

// Error aggregation for batch operations
type MultiError struct {
    Errors []error
}

func (m *MultiError) Error() string {
    if len(m.Errors) == 1 {
        return m.Errors[0].Error()
    }
    return fmt.Sprintf("%d errors occurred", len(m.Errors))
}

func (m *MultiError) Add(err error) {
    if err != nil {
        m.Errors = append(m.Errors, err)
    }
}

func (m *MultiError) HasErrors() bool {
    return len(m.Errors) > 0
}
```

### Error Handling in Functions

```go
// Function that returns multiple values including error
func (l *Library) GetTrack(id string) (*Track, error) {
    l.mu.RLock()
    defer l.mu.RUnlock()
    
    track, exists := l.Tracks[id]
    if !exists {
        return nil, ErrTrackNotFound
    }
    
    return track, nil
}

// Function with validation
func (e *AudioEngine) SetVolume(level float64) error {
    if level < 0 || level > 1 {
        return &ValidationError{
            Field:   "volume",
            Message: fmt.Sprintf("got %.2f, want 0.0-1.0", level),
        }
    }
    
    e.mu.Lock()
    defer e.mu.Unlock()
    e.state.Volume = level
    
    return nil
}

// Defer with error handling
func SaveLibrary(lib *Library, path string) (err error) {
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("create file: %w", err)
    }
    defer func() {
        if cerr := file.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("close file: %w", cerr)
        }
    }()
    
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(lib); err != nil {
        return fmt.Errorf("encode library: %w", err)
    }
    
    return nil
}
```

---

## Maps, Arrays & Slices Usage

### Maps

```go
// Library uses maps for O(1) lookups
type Library struct {
    // Primary index: TrackID -> Track
    Tracks map[string]*Track
    
    // Secondary indices for efficient queries
    artistIndex map[string][]string  // Artist name -> []TrackID
    albumIndex  map[string][]string  // Album name -> []TrackID
    genreIndex  map[string][]string  // Genre -> []TrackID
}

// Initialize maps
func NewLibrary() *Library {
    return &Library{
        Tracks:      make(map[string]*Track),
        artistIndex: make(map[string][]string),
        albumIndex:  make(map[string][]string),
        genreIndex:  make(map[string][]string),
    }
}

// Add track and update indices
func (l *Library) AddTrack(track *Track) {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    l.Tracks[track.ID] = track
    
    // Update indices
    l.artistIndex[track.Artist] = append(l.artistIndex[track.Artist], track.ID)
    l.albumIndex[track.Album] = append(l.albumIndex[track.Album], track.ID)
    l.genreIndex[track.Genre] = append(l.genreIndex[track.Genre], track.ID)
}

// Safe map access with existence check
func (l *Library) GetTracksByArtist(artist string) []*Track {
    l.mu.RLock()
    defer l.mu.RUnlock()
    
    trackIDs, exists := l.artistIndex[artist]
    if !exists {
        return nil
    }
    
    tracks := make([]*Track, 0, len(trackIDs))
    for _, id := range trackIDs {
        if track, ok := l.Tracks[id]; ok {
            tracks = append(tracks, track)
        }
    }
    
    return tracks
}

// Delete from map with index cleanup
func (l *Library) RemoveTrack(id string) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    track, exists := l.Tracks[id]
    if !exists {
        return ErrTrackNotFound
    }
    
    // Remove from indices
    l.removeFromIndex(l.artistIndex, track.Artist, id)
    l.removeFromIndex(l.albumIndex, track.Album, id)
    l.removeFromIndex(l.genreIndex, track.Genre, id)
    
    delete(l.Tracks, id)
    return nil
}
```

### Arrays (Fixed Size)

```go
// Arrays are used when size is known at compile time

// Equalizer bands (fixed 10-band EQ)
type Equalizer struct {
    Bands [10]float64 // Fixed array of 10 frequency bands
}

// Preset EQ configurations
var EQPresets = map[string][10]float64{
    "flat":     {0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
    "bass":     {6, 5, 4, 2, 0, 0, 0, 0, 0, 0},
    "treble":   {0, 0, 0, 0, 0, 2, 4, 5, 6, 6},
    "vocal":    {-2, -1, 0, 2, 4, 4, 2, 0, -1, -2},
}

// History buffer for recently played (circular buffer pattern)
type RecentHistory struct {
    tracks [50]string // Fixed buffer of last 50 tracks
    head   int
    count  int
}

func (h *RecentHistory) Add(trackID string) {
    h.tracks[h.head] = trackID
    h.head = (h.head + 1) % len(h.tracks)
    if h.count < len(h.tracks) {
        h.count++
    }
}
```

### Slices (Dynamic Size)

```go
// Playback queue using slices
type Queue struct {
    tracks []*Track
    index  int
    mu     sync.RWMutex
}

// Append to queue
func (q *Queue) Add(tracks ...*Track) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.tracks = append(q.tracks, tracks...)
}

// Remove from queue (maintaining order)
func (q *Queue) Remove(index int) error {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    if index < 0 || index >= len(q.tracks) {
        return errors.New("index out of bounds")
    }
    
    // Remove element while preserving order
    q.tracks = append(q.tracks[:index], q.tracks[index+1:]...)
    
    // Adjust current index if needed
    if q.index > index {
        q.index--
    }
    
    return nil
}

// Shuffle queue (Fisher-Yates algorithm)
func (q *Queue) Shuffle() {
    q.mu.Lock()
    defer q.mu.Unlock()
    
    n := len(q.tracks)
    for i := n - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i]
    }
    q.index = 0
}

// Slice operations for search results
func (l *Library) Search(query string) []*Track {
    l.mu.RLock()
    defer l.mu.RUnlock()
    
    query = strings.ToLower(query)
    results := make([]*Track, 0, 10) // Pre-allocate with expected capacity
    
    for _, track := range l.Tracks {
        if strings.Contains(strings.ToLower(track.Title), query) ||
           strings.Contains(strings.ToLower(track.Artist), query) {
            results = append(results, track)
        }
    }
    
    // Sort by relevance (title matches first)
    sort.Slice(results, func(i, j int) bool {
        iTitle := strings.Contains(strings.ToLower(results[i].Title), query)
        jTitle := strings.Contains(strings.ToLower(results[j].Title), query)
        return iTitle && !jTitle
    })
    
    return results
}

// Copy slice to prevent modification
func (q *Queue) GetAll() []*Track {
    q.mu.RLock()
    defer q.mu.RUnlock()
    
    result := make([]*Track, len(q.tracks))
    copy(result, q.tracks)
    return result
}
```

---

## Project Structure

```
golang_music_player/
├── cmd/
│   └── player/
│       └── main.go              # Application entry point
│
├── internal/
│   ├── audio/
│   │   ├── engine.go            # Audio playback engine
│   │   ├── engine_test.go       # Engine tests
│   │   ├── decoder.go           # Audio format decoders
│   │   └── effects.go           # Volume, EQ effects
│   │
│   ├── library/
│   │   ├── library.go           # Library management
│   │   ├── library_test.go      # Library tests
│   │   ├── scanner.go           # File scanner
│   │   ├── scanner_test.go      # Scanner tests
│   │   └── metadata.go          # Metadata extraction
│   │
│   ├── playlist/
│   │   ├── playlist.go          # Playlist operations
│   │   ├── playlist_test.go     # Playlist tests
│   │   ├── repository.go        # Playlist storage
│   │   └── queue.go             # Playback queue
│   │
│   ├── config/
│   │   ├── config.go            # Configuration management
│   │   └── config_test.go       # Config tests (JSON marshal/unmarshal)
│   │
│   └── ui/
│       ├── app.go               # Main UI application
│       ├── views/
│       │   ├── player.go        # Player view
│       │   ├── library.go       # Library browser view
│       │   └── playlist.go      # Playlist view
│       └── components/
│           ├── progress.go      # Progress bar
│           ├── list.go          # Track list
│           └── search.go        # Search input
│
├── pkg/
│   ├── events/
│   │   ├── bus.go               # Event bus implementation
│   │   └── types.go             # Event type definitions
│   │
│   └── errors/
│       └── errors.go            # Custom error types
│
├── api/
│   └── types.go                 # Shared types and interfaces
│
├── configs/
│   ├── default.json             # Default configuration
│   └── keybindings.json         # Key binding presets
│
├── testdata/                    # Test fixtures
│   ├── sample.mp3
│   ├── valid_config.json
│   └── invalid_config.json
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── DESIGN.md                    # This document
```

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Set up project structure with `go mod init`
- [ ] Define all core types and interfaces (`api/types.go`)
- [ ] Implement configuration management with JSON (`internal/config/`)
- [ ] Write unit tests for JSON marshal/unmarshal
- [ ] Implement custom error types (`pkg/errors/`)

### Phase 2: Core Engine (Week 2)
- [ ] Implement audio engine with `beep` library
- [ ] Set up goroutine-based playback loop
- [ ] Implement channel-based command/event system
- [ ] Add volume control and seek functionality
- [ ] Write comprehensive tests for audio engine

### Phase 3: Library Management (Week 3)
- [ ] Implement concurrent file scanner with worker pool
- [ ] Add metadata extraction using `tag` library
- [ ] Build in-memory library with map indices
- [ ] Implement search functionality
- [ ] Add playlist CRUD operations with JSON persistence

### Phase 4: Terminal UI (Week 4)
- [ ] Set up `bubbletea` application framework
- [ ] Create player view with playback controls
- [ ] Build library browser view
- [ ] Add playlist management view
- [ ] Implement keyboard navigation and shortcuts

### Phase 5: Polish & Testing (Week 5)
- [ ] Integration testing
- [ ] Error handling improvements
- [ ] Performance optimization
- [ ] Documentation
- [ ] Release preparation

---

## Appendix: Key Dependencies

```go
// go.mod
module github.com/yourusername/golang_music_player

go 1.21

require (
    github.com/faiface/beep v1.1.0       // Audio playback
    github.com/charmbracelet/bubbletea v0.25.0  // TUI framework
    github.com/charmbracelet/lipgloss v0.9.0    // TUI styling
    github.com/dhowden/tag v0.0.0-20230630033851  // Audio metadata
)
```

---

## Quick Reference: Go Concepts Mapping

| Go Concept | Implementation Location |
|------------|------------------------|
| **Goroutines** | Audio engine, Library scanner, Position tracker |
| **Channels** | Command/Event bus, Scanner results, Shutdown signals |
| **Pointers** | All methods modifying state, Map values, Large struct params |
| **Interfaces** | Player, LibraryScanner, Repository, ConfigManager |
| **Structs** | Track, Playlist, Library, Config, PlaybackState |
| **Maps** | Library indices, EQ presets, Key bindings |
| **Slices** | Queue, Search results, Playlist tracks |
| **Arrays** | EQ bands, History buffer |
| **JSON** | Config, Playlists, Library cache |
| **Error handling** | Custom types, Wrapping, Sentinel errors |
| **Unit tests** | Config tests, Playlist tests, Library tests |

---

## Client-Server Architecture

The application has been extended with a standalone HTTP server (`cmd/server/main.go`) that exposes the music library over a REST API. This enables remote access, multi-user authentication, and decouples the audio library from any single client.

### High-Level System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              GTMPC System Architecture                              │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                     │
│   ┌─────────────────────────┐          HTTP/JSON          ┌───────────────────────┐ │
│   │     TUI Client          │ ◄─────────────────────────► │    GTMPC Server       │ │
│   │  (cmd/player/main.go)   │     REST API Calls          │  (cmd/server/main.go) │ │
│   │                         │                             │                       │ │
│   │  ┌───────────────────┐  │                             │  ┌─────────────────┐  │ │
│   │  │  Bubble Tea UI    │  │  POST /api/auth/login       │  │  HTTP Router    │  │ │
│   │  │  ┌─────────────┐  │  │ ──────────────────────────► │  │  (net/http)     │  │ │
│   │  │  │ Player View │  │  │                             │  └────────┬────────┘  │ │
│   │  │  │ Library View│  │  │  GET /api/library/tracks     │           │           │ │
│   │  │  │ Playlist    │  │  │ ──────────────────────────► │  ┌────────▼────────┐  │ │
│   │  │  └─────────────┘  │  │                             │  │   Middleware    │  │ │
│   │  └───────────────────┘  │  GET /api/stream/{id}       │  │  ┌───────────┐  │  │ │
│   │                         │ ──────────────────────────► │  │  │ Logging   │  │  │ │
│   │  ┌───────────────────┐  │       Audio Stream          │  │  │ CORS      │  │  │ │
│   │  │  Audio Engine     │  │ ◄─────────────────────────  │  │  │ JWT Auth  │  │  │ │
│   │  │  (beep/oto)       │  │                             │  │  └───────────┘  │  │ │
│   │  └───────────────────┘  │                             │  └────────┬────────┘  │ │
│   └─────────────────────────┘                             │           │           │ │
│                                                           │  ┌────────▼────────┐  │ │
│                                                           │  │    Handlers     │  │ │
│                                                           │  │  ┌───────────┐  │  │ │
│                                                           │  │  │ Auth      │  │  │ │
│                                                           │  │  │ Library   │  │  │ │
│                                                           │  │  │ Stream    │  │  │ │
│                                                           │  │  └───────────┘  │  │ │
│                                                           │  └────────┬────────┘  │ │
│                                                           │           │           │ │
│                                                           │  ┌────────▼────────┐  │ │
│                                                           │  │   Services      │  │ │
│                                                           │  │  ┌───────────┐  │  │ │
│                                                           │  │  │Auth (bcrypt)│ │  │ │
│                                                           │  │  │Library    │  │  │ │
│                                                           │  │  │Scanner    │  │  │ │
│                                                           │  │  └───────────┘  │  │ │
│                                                           │  └────────┬────────┘  │ │
│                                                           │           │           │ │
│                                                           │  ┌────────▼────────┐  │ │
│                                                           │  │   Storage       │  │ │
│                                                           │  │  users.json     │  │ │
│                                                           │  │  library.json   │  │ │
│                                                           │  │  audio files    │  │ │
│                                                           │  └─────────────────┘  │ │
│                                                           └───────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Server Component Diagram

```
cmd/server/main.go
      │
      ├── internal/config/config.go        (Configuration loading)
      ├── internal/auth/auth.go            (bcrypt user service)
      ├── internal/auth/jwt.go             (HMAC-SHA256 JWT tokens)
      ├── internal/library/library.go      (Music library management)
      └── internal/server/
            ├── server.go                  (Server orchestrator)
            ├── middleware.go              (Logging, CORS, JWT Auth)
            └── handlers.go               (Route handlers)
```

---

## HTTP Server & Routing Design

The server uses Go's standard `net/http` package with `http.ServeMux` for efficient routing and request management.

### Route Table

| Method | Path | Auth | Handler | Description |
|--------|------|------|---------|-------------|
| `GET` | `/api/health` | No | `HandleHealthCheck` | Health probe |
| `POST` | `/api/auth/register` | No | `HandleRegister` | Create user with bcrypt |
| `POST` | `/api/auth/login` | No | `HandleLogin` | Authenticate, return JWT |
| `GET` | `/api/library/tracks` | JWT | `HandleGetTracks` | List all tracks |
| `GET` | `/api/library/search?q=` | JWT | `HandleSearchTracks` | Search tracks |
| `GET` | `/api/stream/{id}` | JWT | `HandleStreamTrack` | Stream audio file |

### Middleware Chain

Requests flow through a composable middleware pipeline:

```
Incoming Request
      │
      ▼
┌─────────────────┐
│ CORS Middleware  │  ← Sets Access-Control headers, handles OPTIONS preflight
└────────┬────────┘
         ▼
┌─────────────────┐
│Logging Middleware│  ← Logs method, path, and latency for every request
└────────┬────────┘
         ▼
┌─────────────────┐
│  JWT Auth       │  ← Validates Bearer token (protected routes only)
│  Middleware     │     Injects X-User and X-Role headers
└────────┬────────┘
         ▼
┌─────────────────┐
│  Route Handler  │  ← Business logic execution
└─────────────────┘
```

### Middleware Implementation Pattern

```go
// Composable middleware chain using functional composition
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Usage: wrapping protected routes
mux.Handle("/api/library/", Chain(protectedMux,
    AuthMiddleware(jwtSecret),
))
```

---

## Security Architecture (bcrypt)

All user credentials are handled using `golang.org/x/crypto/bcrypt` with a cost factor of 12.

### Password Flow

```
┌──────────────┐     ┌──────────────────┐     ┌──────────────────────┐
│  User sends  │     │  bcrypt.Generate │     │  Store hash in       │
│  plaintext   │ ──► │  FromPassword()  │ ──► │  users.json          │
│  password    │     │  (cost=12)       │     │  (never plaintext)   │
└──────────────┘     └──────────────────┘     └──────────────────────┘

┌──────────────┐     ┌──────────────────┐     ┌──────────────────────┐
│  User sends  │     │  bcrypt.Compare  │     │  Return JWT token    │
│  login creds │ ──► │  HashAndPassword │ ──► │  on success, or 401  │
│              │     │  ()              │     │  on failure           │
└──────────────┘     └──────────────────┘     └──────────────────────┘
```

### Key Security Properties

| Property | Implementation |
|----------|----------------|
| **Password hashing** | bcrypt with cost factor 12 |
| **Salt handling** | Automatic per-password salt (built into bcrypt) |
| **Hash never exposed** | `json:"-"` tag on `User.PasswordHash` |
| **Token authentication** | HMAC-SHA256 JWT with configurable TTL |
| **Persistence security** | `users.json` written with `0600` permissions |
| **Thread safety** | `sync.RWMutex` protects concurrent access |

### Auth Service Implementation

```go
// Register hashes the password with bcrypt before storage
func (s *Service) Register(req RegisterRequest) (*User, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
    // ... store hash, never plaintext
}

// Authenticate compares the provided password against the stored hash
func (s *Service) Authenticate(req LoginRequest) (*User, error) {
    err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    // ... return user on success, ErrInvalidPassword on failure
}
```

---

## Server Concurrency Model

The server employs multiple concurrent goroutines for optimized execution:

### Goroutine Architecture

```
Main Goroutine (cmd/server/main.go)
      │
      ├──► Signal Handler Goroutine
      │         Listens for SIGINT/SIGTERM
      │         Triggers context cancellation
      │
      ├──► HTTP Server Goroutine
      │         net/http handles each request in its own goroutine
      │         ├──► Request Goroutine 1 (GET /api/library/tracks)
      │         ├──► Request Goroutine 2 (POST /api/auth/login)
      │         └──► Request Goroutine N (GET /api/stream/{id})
      │
      └──► Background Scanner Goroutine
                Uses ticker for periodic library rescans
                ├──► File Discovery Goroutine (WalkDir)
                └──► Worker Pool Goroutines (4x metadata readers)
```

### Concurrency Patterns Used

| Pattern | Location | Purpose |
|---------|----------|---------|
| **Goroutine-per-request** | `net/http` server | Handle concurrent HTTP clients |
| **Background worker** | `backgroundScanner()` | Periodic library rescanning |
| **Fan-out/Fan-in** | `Scanner.Scan()` | Parallel metadata extraction |
| **Ticker-based scheduling** | `backgroundScanner()` | 5-minute rescan intervals |
| **Context cancellation** | `ctx.Done()` | Graceful shutdown propagation |
| **RWMutex** | `auth.Service`, `Library` | Safe concurrent reads/writes |
| **Buffered channels** | Scanner results | Non-blocking work distribution |

### Background Scanner Implementation

```go
// backgroundScanner runs in its own goroutine with ticker-based scheduling
func (s *Server) backgroundScanner(ctx context.Context, lib *Library, paths []string) {
    // Initial scan at startup
    lib.Scan(ctx, paths)

    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return  // Graceful shutdown
        case <-ticker.C:
            lib.Scan(ctx, paths)  // Periodic rescan
        }
    }
}
```

---

## API Reference

### Register User

```http
POST /api/auth/register
Content-Type: application/json

{
    "username": "alice",
    "password": "securePass123!",
    "role": "user"
}
```

Response (`201 Created`):
```json
{
    "success": true,
    "data": {
        "id": "a1b2c3d4...",
        "username": "alice",
        "role": "user",
        "created_at": "2026-03-27T09:00:00Z"
    }
}
```

### Login

```http
POST /api/auth/login
Content-Type: application/json

{
    "username": "alice",
    "password": "securePass123!"
}
```

Response (`200 OK`):
```json
{
    "success": true,
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": { "id": "a1b2c3d4...", "username": "alice", "role": "user" }
    }
}
```

### Get Library Tracks (Protected)

```http
GET /api/library/tracks
Authorization: Bearer <jwt_token>
```

Response (`200 OK`):
```json
{
    "success": true,
    "data": [
        {
            "id": "abc123",
            "title": "Deep Space",
            "artist": "Aurora",
            "album": "Stellar",
            "duration": 245000000000,
            "genre": "Electronic"
        }
    ]
}
```

### Stream Audio (Protected)

```http
GET /api/stream/{trackId}
Authorization: Bearer <jwt_token>
```

Response: Raw audio bytes with appropriate `Content-Type` (`audio/mpeg`, `audio/wav`, or `audio/flac`). Supports HTTP Range requests for seeking.

---

## PostgreSQL Database Layer

The server uses PostgreSQL as its primary data store, accessed via the `pgx/v5` driver with connection pooling.

### Database Schema

```sql
-- Users table (bcrypt hashed credentials)
CREATE TABLE users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tracks table (music library metadata)
CREATE TABLE tracks (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    artist     TEXT NOT NULL DEFAULT '',
    album      TEXT NOT NULL DEFAULT '',
    duration   BIGINT NOT NULL DEFAULT 0,
    file_path  TEXT NOT NULL,
    genre      TEXT NOT NULL DEFAULT '',
    year       INTEGER NOT NULL DEFAULT 0,
    track_num  INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Playlists table
CREATE TABLE playlists (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Many-to-many junction table
CREATE TABLE playlist_tracks (
    playlist_id TEXT REFERENCES playlists(id) ON DELETE CASCADE,
    track_id    TEXT REFERENCES tracks(id) ON DELETE CASCADE,
    position    INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (playlist_id, track_id)
);

-- Performance indexes
CREATE INDEX idx_tracks_artist ON tracks(artist);
CREATE INDEX idx_tracks_album  ON tracks(album);
CREATE INDEX idx_users_username ON users(username);
```

### Repository Pattern

```
internal/database/
├── database.go       # Connection pool (pgxpool) + auto-migration
├── user_repo.go      # UserRepo: Create, GetByUsername, GetAll
└── track_repo.go     # TrackRepo: Upsert, GetByID, GetAll, Search, GetByArtist
```

| Repository | Key Operations | Concurrency |
|---|---|---|
| `UserRepo` | `Create` (with unique constraint), `GetByUsername`, `GetAll` | Pool handles concurrent requests |
| `TrackRepo` | `Upsert` (idempotent scan sync), `Search` (ILIKE), `GetAll` | Pool handles concurrent requests |

### Connection Pool Configuration

```go
config.MaxConns = 10              // Maximum connections for concurrent HTTP handlers
config.MinConns = 2               // Keep-alive connections for low latency
config.MaxConnLifetime = 30 * min // Rotate connections to handle DNS changes
config.MaxConnIdleTime = 5 * min  // Reclaim idle connections
```

### Data Flow: Scanner → PostgreSQL

```
Background Scanner Goroutine
        │
        ├──► Walk directories (goroutine)
        ├──► Worker Pool: extract metadata (4 goroutines)
        ├──► lib.AddTrack() → in-memory library
        └──► trackRepo.Upsert() → PostgreSQL
                Uses ON CONFLICT DO UPDATE
                for idempotent re-scans
```

---

### Updated Project Structure

```text
gtmpc/
├── api/
│   └── types.go                  # Shared types: Track, User, API responses
├── cmd/
│   ├── player/
│   │   └── main.go               # TUI client entry point
│   └── server/
│       └── main.go               # HTTP server entry point
├── internal/
│   ├── audio/
│   │   ├── engine.go             # Audio playback engine (goroutines)
│   │   └── decoder.go            # MP3/WAV/FLAC decoders
│   ├── auth/
│   │   ├── auth.go               # bcrypt user service (JSON fallback)
│   │   ├── db_service.go         # bcrypt user service (PostgreSQL)
│   │   ├── jwt.go                # HMAC-SHA256 JWT generation/validation
│   │   └── auth_test.go          # Unit tests for auth + JWT
│   ├── config/
│   │   └── config.go             # XDG-compliant configuration
│   ├── database/
│   │   ├── database.go           # pgxpool connection + migrations
│   │   ├── user_repo.go          # User CRUD repository
│   │   └── track_repo.go         # Track CRUD + UPSERT + ILIKE search
│   ├── library/
│   │   ├── library.go            # Library CRUD + search
│   │   ├── scanner.go            # Concurrent file scanner (worker pool)
│   │   └── metadata.go           # ID3/metadata extraction
│   ├── playlist/
│   │   ├── playlist.go           # Playlist manager
│   │   └── queue.go              # Playback queue
│   ├── server/
│   │   ├── server.go             # Server orchestrator + background scanner
│   │   ├── middleware.go         # Logging, CORS, JWT auth middleware
│   │   └── handlers.go           # REST API route handlers
│   └── ui/
│       ├── app.go                # Main Bubble Tea application
│       ├── views/                # Player, Library, Playlist views
│       └── components/           # Progress bar, track list, search
├── pkg/
│   ├── errors/errors.go          # Custom error types
│   └── events/bus.go             # Pub/sub event bus
├── data/
│   ├── library.json              # Persisted track index (local cache)
│   └── playlists/                # Saved playlists
├── go.mod
├── go.sum
├── DESIGN.md                     # This document
├── APPLICATION_FLOW.md           # Execution flow walkthrough
└── README.md                     # Project overview
```

---

*Document Version: 2.0*  
*Last Updated: March 27, 2026*
