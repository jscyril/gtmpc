package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jscyril/golang_music_player/internal/audio"
	"github.com/jscyril/golang_music_player/internal/library"
	"github.com/jscyril/golang_music_player/internal/playlist"
	"github.com/jscyril/golang_music_player/internal/ui/components"
	"github.com/jscyril/golang_music_player/internal/ui/views"
)

func TestLibraryBrowseEnterEmitsFileAddedMsg(t *testing.T) {
	model := NewModel(audio.NewAudioEngine(nil), library.NewLibrary(), playlist.NewManager(t.TempDir()))
	model.activeView = ViewLibrary
	model.libraryView.Browsing = true
	model.libraryView.FileBrowser = components.FileBrowser{
		Entries: []components.FileEntry{
			{Name: "song.mp3", Path: "/tmp/song.mp3", IsDir: false},
		},
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter in file browser to return a command")
	}

	msg := cmd()
	added, ok := msg.(views.FileAddedMsg)
	if !ok {
		t.Fatalf("expected FileAddedMsg, got %T", msg)
	}
	if added.Path != "/tmp/song.mp3" {
		t.Fatalf("expected selected path to be propagated, got %q", added.Path)
	}

	nextModel, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected updated model type, got %T", updated)
	}
	if nextModel.libraryView.Browsing {
		t.Fatal("expected file browser to close after selecting a file")
	}
}
