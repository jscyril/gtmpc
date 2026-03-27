package audio

import (
	"strings"
	"testing"

	"github.com/jscyril/golang_music_player/api"
)

func TestNewAudioEngine(t *testing.T) {
	engine := NewAudioEngine(nil)

	if engine == nil {
		t.Fatal("NewAudioEngine returned nil")
	}

	if engine.state == nil {
		t.Error("Engine state is nil")
	}

	if engine.state.Status != api.StatusStopped {
		t.Errorf("Expected status StatusStopped, got %v", engine.state.Status)
	}

	if engine.state.Volume != 0.5 {
		t.Errorf("Expected volume 0.5, got %f", engine.state.Volume)
	}

	if engine.state.Mode != api.ModeNormal {
		t.Errorf("Expected mode %v, got %v", api.ModeNormal, engine.state.Mode)
	}

	if engine.commands == nil {
		t.Error("Commands channel is nil")
	}

	if engine.events == nil {
		t.Error("Events channel is nil")
	}
}

func TestSetVolume_Valid(t *testing.T) {
	engine := NewAudioEngine(nil)

	tests := []struct {
		name    string
		volume  float64
		wantErr bool
	}{
		{"zero volume", 0.0, false},
		{"half volume", 0.5, false},
		{"full volume", 1.0, false},
		{"below zero", -0.1, true},
		{"above one", 1.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.SetVolume(tt.volume)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetVolume(%f) error = %v, wantErr %v", tt.volume, err, tt.wantErr)
			}
		})
	}
}

func TestGetState(t *testing.T) {
	engine := NewAudioEngine(nil)
	state := engine.GetState()

	if state == nil {
		t.Fatal("GetState returned nil")
	}

	if state.Status != api.StatusStopped {
		t.Errorf("Expected StatusStopped, got %v", state.Status)
	}

	// Verify it's a copy, not the original
	state.Volume = 0.99
	originalState := engine.GetState()
	if originalState.Volume == 0.99 {
		t.Error("GetState should return a copy, not the original")
	}
}

func TestPlay_NilTrack(t *testing.T) {
	engine := NewAudioEngine(nil)

	err := engine.Play(nil)
	if err == nil {
		t.Error("Play(nil) should return an error")
	}
}

func TestSetMode_ValidAndInvalid(t *testing.T) {
	engine := NewAudioEngine(nil)

	if err := engine.SetMode(api.ModeKaraoke); err != nil {
		t.Fatalf("SetMode(karaoke) error = %v", err)
	}

	if err := engine.SetMode(api.AudioMode(99)); err == nil {
		t.Fatal("SetMode(invalid) should return an error")
	}
}

func TestRequestModeChangeMarksSwitchInProgress(t *testing.T) {
	engine := NewAudioEngine(nil)
	engine.state.CurrentTrack = &api.Track{
		ID:       "track-1",
		Title:    "Track 1",
		FilePath: "/tmp/track-1.mp3",
	}
	engine.state.Status = api.StatusPlaying

	if err := engine.requestModeChange(api.ModeKaraoke); err != nil {
		t.Fatalf("requestModeChange(karaoke) error = %v", err)
	}

	state := engine.GetState()
	if !state.ModeSwitching {
		t.Fatal("expected mode switch to be marked as in progress")
	}
	if state.TargetMode != api.ModeKaraoke {
		t.Fatalf("expected target mode %v, got %v", api.ModeKaraoke, state.TargetMode)
	}
	if state.Mode != api.ModeNormal {
		t.Fatalf("expected current mode to remain %v until apply, got %v", api.ModeNormal, state.Mode)
	}
}

func TestFFmpegFilter(t *testing.T) {
	tests := []struct {
		mode    api.AudioMode
		wantErr bool
	}{
		{mode: api.ModeNormal, wantErr: true},
		{mode: api.ModeKaraoke, wantErr: false},
		{mode: api.ModeVocals, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			filter, err := ffmpegFilter(tt.mode)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ffmpegFilter(%s) error = %v, wantErr %v", tt.mode.String(), err, tt.wantErr)
			}
			if !tt.wantErr && filter == "" {
				t.Fatalf("ffmpegFilter(%s) returned an empty filter", tt.mode.String())
			}
			if !tt.wantErr && !strings.Contains(filter, "dialoguenhance") {
				t.Fatalf("ffmpegFilter(%s) should use dialoguenhance, got %q", tt.mode.String(), filter)
			}
		})
	}
}

func TestIsSupported(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/music/song.mp3", true},
		{"/music/song.MP3", true},
		{"/music/song.wav", true},
		{"/music/song.flac", true},
		{"/music/song.ogg", false},
		{"/music/song.aac", false},
		{"/music/song.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsSupported(tt.path)
			if result != tt.expected {
				t.Errorf("IsSupported(%s) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()

	if len(formats) == 0 {
		t.Error("SupportedFormats should return at least one format")
	}

	expected := map[string]bool{".mp3": true, ".wav": true, ".flac": true}
	for _, f := range formats {
		if !expected[f] {
			t.Errorf("Unexpected format: %s", f)
		}
	}
}
