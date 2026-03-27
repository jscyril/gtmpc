package library

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jscyril/golang_music_player/api"
)

func TestPruneMissingTracks(t *testing.T) {
	lib := NewLibrary()

	existingPath := filepath.Join(t.TempDir(), "song.mp3")
	if err := os.WriteFile(existingPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("write existing track fixture: %v", err)
	}

	lib.AddTrack(&api.Track{
		ID:       "exists",
		Title:    "exists",
		FilePath: existingPath,
	})
	lib.AddTrack(&api.Track{
		ID:       "missing",
		Title:    "missing",
		FilePath: filepath.Join(t.TempDir(), "missing.mp3"),
	})

	removed := lib.PruneMissingTracks()
	if removed != 1 {
		t.Fatalf("PruneMissingTracks() removed %d tracks, want 1", removed)
	}

	if _, err := lib.GetTrack("exists"); err != nil {
		t.Fatalf("expected existing track to remain: %v", err)
	}
	if _, err := lib.GetTrack("missing"); err == nil {
		t.Fatal("expected missing track to be removed")
	}
}
