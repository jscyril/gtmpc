// Package screens — player.go implements the now-playing screen with playback controls.
package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/audio"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
)

// BackToLibraryMsg is sent when the user exits the player.
type BackToLibraryMsg struct{}

// NextTrackMsg is sent when the user presses 'n'.
type NextTrackMsg struct{}

// PrevTrackMsg is sent when the user presses 'b'.
type PrevTrackMsg struct{}

// PlayerTickMsg is the tick for progress bar updates.
type PlayerTickMsg time.Time

// PlayerScreen displays the current playback and controls.
type PlayerScreen struct {
	engine    *audio.AudioEngine
	client    *apiclient.APIClient
	track     *apiclient.Track
	allTracks []apiclient.Track
	queueIdx  int
	width     int
	height    int
}

// NewPlayerScreen creates a new player screen for the given track.
func NewPlayerScreen(engine *audio.AudioEngine, client *apiclient.APIClient, track apiclient.Track, all []apiclient.Track, idx int, width, height int) PlayerScreen {
	return PlayerScreen{
		engine:    engine,
		client:    client,
		track:     &track,
		allTracks: all,
		queueIdx:  idx,
		width:     width,
		height:    height,
	}
}

// playerTickCmd produces a tick every 500ms for progress bar updates.
func playerTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return PlayerTickMsg(t)
	})
}

// Init starts the progress bar ticker.
func (s PlayerScreen) Init() tea.Cmd {
	return playerTickCmd()
}

// Update handles player controls.
func (s PlayerScreen) Update(msg tea.Msg) (PlayerScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case PlayerTickMsg:
		return s, playerTickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			state := s.engine.GetState()
			switch state.Status {
			case 1: // StatusPlaying
				s.engine.Pause()
			case 2: // StatusPaused
				s.engine.Resume()
			}
		case "n":
			return s, func() tea.Msg { return NextTrackMsg{} }
		case "b":
			return s, func() tea.Msg { return PrevTrackMsg{} }
		case "+", "=":
			state := s.engine.GetState()
			v := state.Volume + 0.1
			if v > 1 {
				v = 1
			}
			s.engine.SetVolume(v)
		case "-":
			state := s.engine.GetState()
			v := state.Volume - 0.1
			if v < 0 {
				v = 0
			}
			s.engine.SetVolume(v)
		case "right":
			state := s.engine.GetState()
			s.engine.Seek(state.Position + 5*time.Second)
		case "left":
			state := s.engine.GetState()
			newPos := state.Position - 5*time.Second
			if newPos < 0 {
				newPos = 0
			}
			s.engine.Seek(newPos)
		case "q", "esc":
			return s, func() tea.Msg { return BackToLibraryMsg{} }
		}
	}

	return s, nil
}

// View renders the player screen.
func (s PlayerScreen) View() string {
	if s.track == nil {
		return styles.SubtitleStyle.Render("No track selected")
	}

	state := s.engine.GetState()

	var sb strings.Builder

	// Status icon
	statusIcon := "⏹"
	switch state.Status {
	case 1: // Playing
		statusIcon = "▶"
	case 2: // Paused
		statusIcon = "⏸"
	}

	sb.WriteString("\n")
	sb.WriteString(styles.TitleStyle.Render(statusIcon+" "+s.track.Title) + "\n\n")
	sb.WriteString(styles.NowPlayingStyle.Render(s.track.Artist) + "\n")
	sb.WriteString(styles.SubtitleStyle.Render(s.track.Album) + "\n\n")

	// Cover art placeholder
	coverBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ColorBorder).
		Width(20).
		Height(5).
		Align(lipgloss.Center).
		Render("[ Cover Art ]")
	sb.WriteString(coverBox + "\n\n")

	// Progress bar
	pos := state.Position
	total := time.Duration(s.track.DurationSeconds) * time.Second
	progress := renderProgressBar(pos, total, s.width-8)
	sb.WriteString(progress + "\n")
	sb.WriteString(styles.SubtitleStyle.Render(
		fmt.Sprintf("  %s / %s", formatDur(pos), formatDur(total)),
	) + "\n\n")

	// Volume
	sb.WriteString(fmt.Sprintf("Volume: %s %d%%\n\n",
		renderVolBar(state.Volume),
		int(state.Volume*100),
	))

	// Queue position
	if len(s.allTracks) > 0 {
		sb.WriteString(styles.SubtitleStyle.Render(
			fmt.Sprintf("Track %d / %d in queue", s.queueIdx+1, len(s.allTracks)),
		) + "\n\n")
	}

	sb.WriteString(styles.HelpStyle.Render(
		"[Space] Play/Pause  [n] Next  [b] Prev  [+/-] Volume  [←/→] Seek ±5s  [Esc/q] Back",
	))

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 3).
		Width(s.width - 4).
		Render(sb.String())

	return card
}

// renderProgressBar renders a text-based progress bar.
func renderProgressBar(pos, total time.Duration, width int) string {
	if width < 4 {
		width = 4
	}
	if total <= 0 {
		return styles.ProgressBarEmptyStyle.Render(strings.Repeat("─", width))
	}
	ratio := float64(pos) / float64(total)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	empty := width - filled

	bar := styles.ProgressBarStyle.Render(strings.Repeat("█", filled)) +
		styles.ProgressBarEmptyStyle.Render(strings.Repeat("─", empty))
	return bar
}

// renderVolBar renders a compact volume indicator.
func renderVolBar(vol float64) string {
	filled := int(vol * 10)
	empty := 10 - filled
	return styles.ProgressBarStyle.Render(strings.Repeat("●", filled)) +
		styles.ProgressBarEmptyStyle.Render(strings.Repeat("○", empty))
}

// formatDur formats a time.Duration as m:ss.
func formatDur(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	m := int(d.Minutes())
	sec := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, sec)
}
