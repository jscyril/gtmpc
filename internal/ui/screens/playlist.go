// Package screens — playlist.go implements the playlist browsing and management screen.
package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
)

// PlaylistScreen displays playlists and their tracks.
type PlaylistScreen struct {
	client      *apiclient.APIClient
	playlists   []apiclient.Playlist
	trackMap    map[string]apiclient.Track // track id -> track
	selected    int
	trackSel    int
	showTracks  bool
	loading     bool
	err         string
	spinner     spinner.Model
	width       int
	height      int
}

type playlistDataMsg struct {
	playlists []apiclient.Playlist
	tracks    []apiclient.Track
	err       error
}

// NewPlaylistScreen creates a new PlaylistScreen.
func NewPlaylistScreen(client *apiclient.APIClient, width, height int) PlaylistScreen {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorPrimary)

	return PlaylistScreen{
		client:   client,
		loading:  true,
		spinner:  sp,
		trackMap: make(map[string]apiclient.Track),
		width:    width,
		height:   height,
	}
}

// fetchData fetches both playlists and tracks from the API.
func (s PlaylistScreen) fetchData() tea.Cmd {
	return func() tea.Msg {
		pResp, err := s.client.GetPlaylists()
		if err != nil {
			return playlistDataMsg{err: err}
		}
		tResp, err := s.client.GetTracks()
		if err != nil {
			return playlistDataMsg{playlists: pResp.Playlists, err: err}
		}
		return playlistDataMsg{playlists: pResp.Playlists, tracks: tResp.Tracks}
	}
}

// Init starts data fetch.
func (s PlaylistScreen) Init() tea.Cmd {
	return tea.Batch(s.fetchData(), s.spinner.Tick)
}

// Update handles messages.
func (s PlaylistScreen) Update(msg tea.Msg) (PlaylistScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case playlistDataMsg:
		s.loading = false
		if msg.err != nil {
			s.err = humanizeError(msg.err)
		} else {
			s.playlists = msg.playlists
			for _, t := range msg.tracks {
				s.trackMap[t.ID] = t
			}
		}

	case tea.KeyMsg:
		if s.loading {
			return s, nil
		}

		if s.showTracks {
			switch msg.String() {
			case "esc", "backspace", "b":
				s.showTracks = false
				s.trackSel = 0
			case "j", "down":
				if s.selected < len(s.playlists) {
					pl := s.playlists[s.selected]
					if s.trackSel < len(pl.TrackIDs)-1 {
						s.trackSel++
					}
				}
			case "k", "up":
				if s.trackSel > 0 {
					s.trackSel--
				}
			case "enter":
				if s.selected < len(s.playlists) {
					pl := s.playlists[s.selected]
					if s.trackSel < len(pl.TrackIDs) {
						trackID := pl.TrackIDs[s.trackSel]
						if track, ok := s.trackMap[trackID]; ok {
							// Build the slice of all tracks in this playlist
							allTracks := make([]apiclient.Track, 0, len(pl.TrackIDs))
							for _, id := range pl.TrackIDs {
								if t, ok := s.trackMap[id]; ok {
									allTracks = append(allTracks, t)
								}
							}
							idx := s.trackSel
							return s, func() tea.Msg {
								return PlayTrackMsg{Track: track, AllTracks: allTracks, Index: idx}
							}
						}
					}
				}
			}
		} else {
			switch msg.String() {
			case "j", "down":
				if s.selected < len(s.playlists)-1 {
					s.selected++
				}
			case "k", "up":
				if s.selected > 0 {
					s.selected--
				}
			case "enter":
				if len(s.playlists) > 0 && s.selected < len(s.playlists) {
					s.showTracks = true
					s.trackSel = 0
				}
			case "esc", "q":
				return s, func() tea.Msg { return BackToLibraryMsg{} }
			}
		}
	}

	return s, nil
}

// View renders the playlist screen.
func (s PlaylistScreen) View() string {
	if s.loading {
		return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center,
			styles.TitleStyle.Render(s.spinner.View()+" Loading playlists..."))
	}

	var sb strings.Builder

	if s.err != "" {
		sb.WriteString(styles.ErrorStyle.Render("✗ "+s.err) + "\n\n")
	}

	if s.showTracks && s.selected < len(s.playlists) {
		pl := s.playlists[s.selected]
		sb.WriteString(styles.TitleStyle.Render("📋 "+pl.Name) + "\n")
		sb.WriteString(styles.SubtitleStyle.Render(fmt.Sprintf("%d tracks", len(pl.TrackIDs))) + "\n\n")

		if len(pl.TrackIDs) == 0 {
			sb.WriteString(styles.SubtitleStyle.Render("This playlist is empty.") + "\n")
		} else {
			for i, id := range pl.TrackIDs {
				t, ok := s.trackMap[id]
				var row string
				if ok {
					row = fmt.Sprintf("%-4d %-30s %-20s", i+1, truncate(t.Title, 30), truncate(t.Artist, 20))
				} else {
					row = fmt.Sprintf("%-4d %-30s", i+1, "(unknown track)")
				}
				if i == s.trackSel {
					sb.WriteString(styles.SelectedRowStyle.Render(row) + "\n")
				} else {
					sb.WriteString(row + "\n")
				}
			}
		}

		sb.WriteString("\n")
		sb.WriteString(styles.HelpStyle.Render("[j/k] Navigate  [Enter] Play  [Esc/b] Back to Playlists"))
	} else {
		sb.WriteString(styles.TitleStyle.Render("📋 Playlists") + "\n\n")

		if len(s.playlists) == 0 {
			sb.WriteString(styles.SubtitleStyle.Render("No playlists found.") + "\n")
		} else {
			for i, pl := range s.playlists {
				row := fmt.Sprintf("%-30s  (%d tracks)", pl.Name, len(pl.TrackIDs))
				if i == s.selected {
					sb.WriteString(styles.SelectedRowStyle.Render(row) + "\n")
				} else {
					sb.WriteString(row + "\n")
				}
			}
		}

		sb.WriteString("\n")
		sb.WriteString(styles.HelpStyle.Render("[j/k] Navigate  [Enter] Open  [Esc/q] Back to Library"))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(s.width - 4).
		Render(sb.String())
}
