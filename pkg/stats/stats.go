// Package stats provides in-memory session statistics for the gtmpc music player.
// It tracks play events, liked tracks, and computes statistical measures such as
// mean track duration and standard deviation — demonstrating advanced numeric
// processing and string manipulation in Go.
package stats

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// PlayEvent records a single track-play occurrence.
type PlayEvent struct {
	TrackID      string
	Title        string
	Artist       string
	Album        string
	DurationSecs int
	PlayedAt     time.Time
}

// StatsSummary is a fully computed snapshot of the current session statistics.
// It includes derived mathematical values (mean, standard deviation) as well
// as formatted string representations for display.
type StatsSummary struct {
	// Counts
	TracksPlayed int
	TracksLiked  int

	// Time
	TotalSeconds  int
	FormattedTime string // e.g. "1h 24m 08s"

	// Top artist (string manipulation: frequency map + sort)
	TopArtist        string
	ArtistPlayCounts map[string]int // artist → play count (sorted by count desc)

	// Mathematical computations
	MeanDurationSec float64 // arithmetic mean of all played track durations
	StdDevSec       float64 // population standard deviation

	// Most-played track
	MostPlayedTitle  string
	MostPlayedCount  int

	// Formatted stats for display
	FormattedMean   string // e.g. "3:42"
	FormattedStdDev string // e.g. "±0:28"

	// Artist bar chart data (sorted desc by count)
	ArtistChart []ArtistBar
}

// ArtistBar represents one row in the artist breakdown chart.
type ArtistBar struct {
	Artist string
	Count  int
	Bar    string // pre-rendered ASCII bar e.g. "████████"
}

// Stats is the thread-safe in-memory statistics tracker.
type Stats struct {
	mu     sync.RWMutex
	events []PlayEvent
	likes  map[string]bool // trackID → liked
}

// New creates an initialised Stats tracker.
func New() *Stats {
	return &Stats{
		likes: make(map[string]bool),
	}
}

// RecordPlay records a track play event. Safe to call from any goroutine.
func (s *Stats) RecordPlay(trackID, title, artist, album string, durationSecs int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, PlayEvent{
		TrackID:      trackID,
		Title:        title,
		Artist:       artist,
		Album:        album,
		DurationSecs: durationSecs,
		PlayedAt:     time.Now(),
	})
}

// ToggleLike toggles the liked state of a track and returns the new liked state.
func (s *Stats) ToggleLike(trackID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.likes[trackID] = !s.likes[trackID]
	return s.likes[trackID]
}

// IsLiked returns whether a track is currently liked.
func (s *Stats) IsLiked(trackID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.likes[trackID]
}

// Clear resets all statistics for the current session.
func (s *Stats) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = nil
	s.likes = make(map[string]bool)
}

// Summary computes and returns the full StatsSummary from the recorded events.
// All mathematical computations are performed here.
func (s *Stats) Summary() StatsSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sum := StatsSummary{
		ArtistPlayCounts: make(map[string]int),
	}

	if len(s.events) == 0 {
		sum.FormattedTime = "0s"
		sum.FormattedMean = "—"
		sum.FormattedStdDev = "—"
		sum.TracksLiked = countTrue(s.likes)
		return sum
	}

	// ── Basic counts ───────────────────────────────────────────────────────
	sum.TracksPlayed = len(s.events)
	sum.TracksLiked = countTrue(s.likes)

	// ── Total listening time ───────────────────────────────────────────────
	totalSecs := 0
	for _, e := range s.events {
		totalSecs += e.DurationSecs
	}
	sum.TotalSeconds = totalSecs
	sum.FormattedTime = FormatListenTime(totalSecs)

	// ── Artist frequency map (string key, int value) ───────────────────────
	for _, e := range s.events {
		artist := strings.TrimSpace(e.Artist)
		if artist == "" {
			artist = "Unknown"
		}
		sum.ArtistPlayCounts[artist]++
	}

	// Find top artist by frequency
	topCount := 0
	for artist, count := range sum.ArtistPlayCounts {
		if count > topCount || (count == topCount && artist < sum.TopArtist) {
			topCount = count
			sum.TopArtist = artist
		}
	}

	// ── Most-played track (extra string manipulation) ───────────────────────
	trackCounts := make(map[string]int)
	trackTitles := make(map[string]string) // id → title
	for _, e := range s.events {
		trackCounts[e.TrackID]++
		trackTitles[e.TrackID] = e.Title
	}
	for id, count := range trackCounts {
		if count > sum.MostPlayedCount {
			sum.MostPlayedCount = count
			sum.MostPlayedTitle = truncate(trackTitles[id], 30)
		}
	}

	// ── Mathematical: mean duration (arithmetic mean) ──────────────────────
	// mean = Σ(duration_i) / n
	n := float64(len(s.events))
	sumDur := 0.0
	for _, e := range s.events {
		sumDur += float64(e.DurationSecs)
	}
	mean := sumDur / n
	sum.MeanDurationSec = mean
	sum.FormattedMean = formatSecsDuration(int(math.Round(mean)))

	// ── Mathematical: population standard deviation ────────────────────────
	// σ = sqrt( Σ(xi - μ)² / n )
	variance := 0.0
	for _, e := range s.events {
		diff := float64(e.DurationSecs) - mean
		variance += diff * diff
	}
	variance /= n
	stddev := math.Sqrt(variance)
	sum.StdDevSec = stddev
	sum.FormattedStdDev = "±" + formatSecsDuration(int(math.Round(stddev)))

	// ── Artist bar chart (sorted by count desc) ────────────────────────────
	type kv struct {
		Artist string
		Count  int
	}
	var sorted []kv
	for a, c := range sum.ArtistPlayCounts {
		sorted = append(sorted, kv{a, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Count != sorted[j].Count {
			return sorted[i].Count > sorted[j].Count
		}
		return sorted[i].Artist < sorted[j].Artist
	})

	maxCount := 0
	if len(sorted) > 0 {
		maxCount = sorted[0].Count
	}
	barWidth := 12
	for i, kv := range sorted {
		if i >= 8 { // cap chart at 8 entries
			break
		}
		filled := barWidth
		if maxCount > 0 {
			filled = int(math.Round(float64(kv.Count) / float64(maxCount) * float64(barWidth)))
		}
		if filled < 1 && kv.Count > 0 {
			filled = 1
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		sum.ArtistChart = append(sum.ArtistChart, ArtistBar{
			Artist: truncate(kv.Artist, 20),
			Count:  kv.Count,
			Bar:    bar,
		})
	}

	return sum
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// FormatListenTime converts total seconds into a human-readable duration string.
// Examples: 0 → "0s", 45 → "45s", 90 → "1m 30s", 3725 → "1h 02m 05s"
// This demonstrates advanced string manipulation: conditional unit rendering.
func FormatListenTime(totalSecs int) string {
	if totalSecs <= 0 {
		return "0s"
	}
	h := totalSecs / 3600
	m := (totalSecs % 3600) / 60
	sec := totalSecs % 60

	var parts []string
	if h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
		parts = append(parts, fmt.Sprintf("%02dm", m))
		parts = append(parts, fmt.Sprintf("%02ds", sec))
	} else if m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
		parts = append(parts, fmt.Sprintf("%02ds", sec))
	} else {
		parts = append(parts, fmt.Sprintf("%ds", sec))
	}
	return strings.Join(parts, " ")
}

// formatSecsDuration formats integer seconds as "m:ss" for display.
func formatSecsDuration(secs int) string {
	if secs < 0 {
		secs = 0
	}
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

// truncate shortens s to max runes, appending "…" if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max < 2 {
		return string(runes[:max])
	}
	return string(runes[:max-1]) + "…"
}

// countTrue counts how many values in m are true.
func countTrue(m map[string]bool) int {
	n := 0
	for _, v := range m {
		if v {
			n++
		}
	}
	return n
}
