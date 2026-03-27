// Package screens — stats.go implements the session statistics screen.
// It displays aggregated playback data: songs played, liked, total listening
// time, top artist, mean/stddev duration, and an artist bar chart.
package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/stats"
)

// StatsScreen renders the session statistics view.
type StatsScreen struct {
	stats  *stats.Stats
	width  int
	height int
}

// NewStatsScreen creates a new StatsScreen backed by the given Stats tracker.
func NewStatsScreen(s *stats.Stats, width, height int) StatsScreen {
	return StatsScreen{stats: s, width: width, height: height}
}

// Init is a no-op (stats are computed live from memory).
func (s StatsScreen) Init() tea.Cmd { return nil }

// Update handles key input for the stats screen.
func (s StatsScreen) Update(msg tea.Msg) (StatsScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return s, func() tea.Msg { return BackToLibraryMsg{} }
		case "l", "c":
			// [l] or [c] clears the current session stats
			s.stats.Clear()
		}
	}
	return s, nil
}

// View renders the statistics screen.
func (s StatsScreen) View() string {
	sum := s.stats.Summary()

	var sb strings.Builder

	// ── Title ──────────────────────────────────────────────────────────────
	sb.WriteString(styles.TitleStyle.Render("📊 Session Statistics") + "\n\n")

	if sum.TracksPlayed == 0 {
		sb.WriteString(styles.SubtitleStyle.Render("No tracks played yet this session.\nStart listening to see your stats!") + "\n\n")
		sb.WriteString(styles.HelpStyle.Render("[Esc/q] Back"))
		return wrapInCard(sb.String(), s.width)
	}

	// ── Stat rows (label + value) ──────────────────────────────────────────
	rows := []struct{ label, value string }{
		{"Songs Played", fmt.Sprintf("%d", sum.TracksPlayed)},
		{"Songs Liked", fmt.Sprintf("♥  %d", sum.TracksLiked)},
		{"Listen Time", sum.FormattedTime},
		{"Top Artist", emptyOr(sum.TopArtist, "—")},
		{"Avg Duration", sum.FormattedMean},
		{"Std Deviation", sum.FormattedStdDev},
	}
	if sum.MostPlayedTitle != "" && sum.MostPlayedCount > 1 {
		rows = append(rows, struct{ label, value string }{
			"Most Replayed", fmt.Sprintf("%s (%dx)", sum.MostPlayedTitle, sum.MostPlayedCount),
		})
	}

	labelStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted).Width(16)
	valueStyle := lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true)

	for _, r := range rows {
		sb.WriteString(labelStyle.Render(r.label) + "  " + valueStyle.Render(r.value) + "\n")
	}
	sb.WriteString("\n")

	// ── Artist breakdown chart ─────────────────────────────────────────────
	if len(sum.ArtistChart) > 0 {
		sb.WriteString(styles.TableHeaderStyle.Render("Artist Breakdown") + "\n")
		artistLabel := lipgloss.NewStyle().Foreground(styles.ColorMuted).Width(22)
		barStyle := lipgloss.NewStyle().Foreground(styles.ColorPrimary)
		countStyle := lipgloss.NewStyle().Foreground(styles.ColorSecondary)

		for _, bar := range sum.ArtistChart {
			line := artistLabel.Render(bar.Artist) + "  " +
				barStyle.Render(bar.Bar) + "  " +
				countStyle.Render(fmt.Sprintf("%d", bar.Count))
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	// ── Help bar ───────────────────────────────────────────────────────────
	sb.WriteString(styles.HelpStyle.Render("[Esc/q] Back  [c] Clear session"))

	return wrapInCard(sb.String(), s.width)
}

// wrapInCard wraps the content in a rounded border card centered on screen.
func wrapInCard(content string, width int) string {
	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 3).
		Width(cardWidth).
		Render(content)
}

// emptyOr returns fallback if s is empty.
func emptyOr(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}
