// Package stats_test tests the stats package computational and string formatting logic.
package stats

import (
	"math"
	"strings"
	"testing"
)

// TestRecordPlay verifies that RecordPlay increments the play count correctly.
func TestRecordPlay(t *testing.T) {
	s := New()

	if got := len(s.events); got != 0 {
		t.Fatalf("expected 0 events initially, got %d", got)
	}

	s.RecordPlay("id1", "Song A", "Artist X", "Album 1", 200)
	s.RecordPlay("id2", "Song B", "Artist Y", "Album 2", 180)
	s.RecordPlay("id1", "Song A", "Artist X", "Album 1", 200) // replay

	if got := len(s.events); got != 3 {
		t.Errorf("expected 3 events after 3 RecordPlay calls, got %d", got)
	}

	sum := s.Summary()
	if sum.TracksPlayed != 3 {
		t.Errorf("Summary().TracksPlayed = %d, want 3", sum.TracksPlayed)
	}
}

// TestToggleLike verifies like toggling and idempotency.
func TestToggleLike(t *testing.T) {
	s := New()

	// Initially not liked
	if s.IsLiked("id1") {
		t.Error("expected id1 to not be liked initially")
	}

	// First toggle → liked
	liked := s.ToggleLike("id1")
	if !liked {
		t.Error("ToggleLike should return true after first toggle")
	}
	if !s.IsLiked("id1") {
		t.Error("expected id1 to be liked after toggle")
	}

	// Second toggle → unliked
	liked = s.ToggleLike("id1")
	if liked {
		t.Error("ToggleLike should return false after second toggle")
	}
	if s.IsLiked("id1") {
		t.Error("expected id1 to not be liked after second toggle")
	}

	// Summary reflects likes count correctly
	s.ToggleLike("id1") // like
	s.ToggleLike("id2") // like
	sum := s.Summary()
	if sum.TracksLiked != 2 {
		t.Errorf("Summary().TracksLiked = %d, want 2", sum.TracksLiked)
	}
}

// TestSummaryMath verifies mean and standard deviation computations using
// a deterministic dataset where the expected values can be calculated by hand.
//
// Dataset: durations [180, 240, 300]
//   mean = (180+240+300)/3 = 240s
//   variance = ((180-240)² + (240-240)² + (300-240)²) / 3
//            = (3600 + 0 + 3600) / 3 = 2400
//   stddev = √2400 ≈ 48.99
func TestSummaryMath(t *testing.T) {
	s := New()
	s.RecordPlay("a", "Track A", "Artist", "Album", 180)
	s.RecordPlay("b", "Track B", "Artist", "Album", 240)
	s.RecordPlay("c", "Track C", "Artist", "Album", 300)

	sum := s.Summary()

	// Mean
	wantMean := 240.0
	if math.Abs(sum.MeanDurationSec-wantMean) > 0.01 {
		t.Errorf("MeanDurationSec = %.4f, want %.4f", sum.MeanDurationSec, wantMean)
	}

	// Standard deviation
	wantStdDev := math.Sqrt(2400)
	if math.Abs(sum.StdDevSec-wantStdDev) > 0.01 {
		t.Errorf("StdDevSec = %.4f, want %.4f", sum.StdDevSec, wantStdDev)
	}

	// Total seconds
	if sum.TotalSeconds != 720 {
		t.Errorf("TotalSeconds = %d, want 720", sum.TotalSeconds)
	}

	// Formatted mean: 240s = 4:00
	if sum.FormattedMean != "4:00" {
		t.Errorf("FormattedMean = %q, want %q", sum.FormattedMean, "4:00")
	}
}

// TestTopArtist verifies that the frequency map and top-artist derivation are correct.
func TestTopArtist(t *testing.T) {
	s := New()
	s.RecordPlay("1", "Song", "Beatles", "Abbey Road", 200)
	s.RecordPlay("2", "Song", "Beatles", "Let It Be", 220)
	s.RecordPlay("3", "Song", "Beatles", "Revolver", 180)
	s.RecordPlay("4", "Song", "Radiohead", "OK Computer", 240)
	s.RecordPlay("5", "Song", "Radiohead", "Kid A", 260)

	sum := s.Summary()

	if sum.TopArtist != "Beatles" {
		t.Errorf("TopArtist = %q, want %q", sum.TopArtist, "Beatles")
	}

	if got := sum.ArtistPlayCounts["Beatles"]; got != 3 {
		t.Errorf("ArtistPlayCounts[Beatles] = %d, want 3", got)
	}
	if got := sum.ArtistPlayCounts["Radiohead"]; got != 2 {
		t.Errorf("ArtistPlayCounts[Radiohead] = %d, want 2", got)
	}

	// Chart should be ordered desc
	if len(sum.ArtistChart) < 2 {
		t.Fatal("expected at least 2 entries in artist chart")
	}
	if sum.ArtistChart[0].Artist != "Beatles" {
		t.Errorf("ArtistChart[0] = %q, want Beatles", sum.ArtistChart[0].Artist)
	}
}

// TestFormatListenTime verifies string-formatting of various total-second values.
func TestFormatListenTime(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{0, "0s"},
		{-5, "0s"},
		{30, "30s"},
		{60, "1m 00s"},
		{90, "1m 30s"},
		{3600, "1h 00m 00s"},
		{3725, "1h 02m 05s"},
		{7384, "2h 03m 04s"},
	}

	for _, tt := range tests {
		got := FormatListenTime(tt.secs)
		if got != tt.want {
			t.Errorf("FormatListenTime(%d) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

// TestClearResetsState verifies Clear() resets all events and likes.
func TestClearResetsState(t *testing.T) {
	s := New()
	s.RecordPlay("1", "Song", "Artist", "Album", 200)
	s.ToggleLike("1")
	s.Clear()

	sum := s.Summary()
	if sum.TracksPlayed != 0 {
		t.Errorf("after Clear, TracksPlayed = %d, want 0", sum.TracksPlayed)
	}
	if sum.TracksLiked != 0 {
		t.Errorf("after Clear, TracksLiked = %d, want 0", sum.TracksLiked)
	}
	if !strings.Contains(sum.FormattedTime, "0s") {
		t.Errorf("after Clear, FormattedTime = %q, want '0s'", sum.FormattedTime)
	}
}
