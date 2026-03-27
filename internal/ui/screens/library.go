// Package screens — library.go implements the music library browsing screen.
// It fetches tracks from the server API and displays them in a scrollable table.
package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
)

// PlayTrackMsg is sent when the user selects a track to play.
type PlayTrackMsg struct {
	Track     apiclient.Track
	AllTracks []apiclient.Track
	Index     int
}

// GoToPlaylistMsg is sent when the user presses 'p'.
type GoToPlaylistMsg struct{}

// LibraryScreen displays the track library fetched from the server.
type LibraryScreen struct {
	client    *apiclient.APIClient
	allTracks []apiclient.Track
	filtered  []apiclient.Track
	selected  int
	loading   bool
	err       string
	searching bool
	search    textinput.Model
	spinner   spinner.Model
	width     int
	height    int
}

// libraryTracksMsg carries the result of fetching tracks from the API.
type libraryTracksMsg struct {
	tracks []apiclient.Track
	err    error
}

// NewLibraryScreen creates a new LibraryScreen.
func NewLibraryScreen(client *apiclient.APIClient, width, height int) LibraryScreen {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.ColorPrimary)

	srch := textinput.New()
	srch.Placeholder = "Search tracks..."
	srch.Width = 40

	return LibraryScreen{
		client:  client,
		loading: true,
		spinner: sp,
		search:  srch,
		width:   width,
		height:  height,
	}
}

// fetchTracks returns a command that calls the API.
func (s LibraryScreen) fetchTracks() tea.Cmd {
	return func() tea.Msg {
		resp, err := s.client.GetTracks()
		if err != nil {
			return libraryTracksMsg{err: err}
		}
		return libraryTracksMsg{tracks: resp.Tracks}
	}
}

// Init triggers initial track fetch and starts the spinner.
func (s LibraryScreen) Init() tea.Cmd {
	return tea.Batch(s.fetchTracks(), s.spinner.Tick)
}

// Update handles messages.
func (s LibraryScreen) Update(msg tea.Msg) (LibraryScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd

	case libraryTracksMsg:
		s.loading = false
		if msg.err != nil {
			s.err = humanizeError(msg.err)
		} else {
			s.allTracks = msg.tracks
			s.filtered = msg.tracks
			s.selected = 0
		}

	case tea.KeyMsg:
		if s.loading {
			return s, nil
		}

		if s.searching {
			switch msg.String() {
			case "esc", "enter":
				s.searching = false
				s.search.Blur()
			default:
				var cmd tea.Cmd
				s.search, cmd = s.search.Update(msg)
				s.filterTracks(s.search.Value())
				return s, cmd
			}
			return s, nil
		}

		switch msg.String() {
		case "j", "down":
			if s.selected < len(s.filtered)-1 {
				s.selected++
			}
		case "k", "up":
			if s.selected > 0 {
				s.selected--
			}
		case "enter":
			if len(s.filtered) > 0 && s.selected < len(s.filtered) {
				track := s.filtered[s.selected]
				return s, func() tea.Msg {
					return PlayTrackMsg{
						Track:     track,
						AllTracks: s.filtered,
						Index:     s.selected,
					}
				}
			}
		case "/":
			s.searching = true
			s.search.Focus()
		case "p":
			return s, func() tea.Msg { return GoToPlaylistMsg{} }
		case "r":
			s.loading = true
			s.err = ""
			return s, tea.Batch(s.fetchTracks(), s.spinner.Tick)
		}
	}

	return s, nil
}

// filterTracks filters the track list by title, artist, or album.
func (s *LibraryScreen) filterTracks(query string) {
	if query == "" {
		s.filtered = s.allTracks
		s.selected = 0
		return
	}
	q := strings.ToLower(query)
	s.filtered = nil
	for _, t := range s.allTracks {
		if strings.Contains(strings.ToLower(t.Title), q) ||
			strings.Contains(strings.ToLower(t.Artist), q) ||
			strings.Contains(strings.ToLower(t.Album), q) {
			s.filtered = append(s.filtered, t)
		}
	}
	s.selected = 0
}

// formatDuration formats seconds as m:ss
func formatDuration(secs int) string {
	d := time.Duration(secs) * time.Second
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

// View renders the library screen.
func (s LibraryScreen) View() string {
	if s.loading {
		return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center,
			styles.TitleStyle.Render(s.spinner.View()+" Loading library..."))
	}

	var sb strings.Builder

	// Header
	header := styles.TitleStyle.Render("🎵 Music Library")
	count := styles.SubtitleStyle.Render(fmt.Sprintf("  %d tracks", len(s.filtered)))
	sb.WriteString(header + count + "\n\n")

	// Search bar
	if s.searching {
		sb.WriteString("Search: " + s.search.View() + "\n\n")
	} else {
		sb.WriteString(styles.HelpStyle.Render("[/] to search") + "\n\n")
	}

	// Error display
	if s.err != "" {
		sb.WriteString(styles.ErrorStyle.Render("✗ "+s.err) + "\n\n")
	}

	// Table header
	colWidths := libraryColWidths(s.width)
	header2 := fmt.Sprintf("%-4s %-*s %-*s %-*s %s",
		"#",
		colWidths[0], "TITLE",
		colWidths[1], "ARTIST",
		colWidths[2], "ALBUM",
		"TIME",
	)
	sb.WriteString(styles.TableHeaderStyle.Render(header2) + "\n")

	// Table rows — show a window of rows around the selection
	visibleRows := s.height - 12
	if visibleRows < 1 {
		visibleRows = 10
	}
	start := 0
	if s.selected >= visibleRows {
		start = s.selected - visibleRows + 1
	}
	end := start + visibleRows
	if end > len(s.filtered) {
		end = len(s.filtered)
	}

	for i := start; i < end; i++ {
		t := s.filtered[i]
		title := truncate(t.Title, colWidths[0])
		artist := truncate(t.Artist, colWidths[1])
		album := truncate(t.Album, colWidths[2])
		dur := formatDuration(t.DurationSeconds)
		row := fmt.Sprintf("%-4d %-*s %-*s %-*s %s",
			i+1,
			colWidths[0], title,
			colWidths[1], artist,
			colWidths[2], album,
			dur,
		)
		if i == s.selected {
			sb.WriteString(styles.SelectedRowStyle.Render(row) + "\n")
		} else {
			sb.WriteString(row + "\n")
		}
	}

	if len(s.filtered) == 0 && !s.loading {
		if s.search.Value() != "" {
			sb.WriteString(styles.SubtitleStyle.Render("No tracks match your search.") + "\n")
		} else {
			sb.WriteString(styles.SubtitleStyle.Render("No tracks found. Add music to the server.") + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(styles.HelpStyle.Render("[j/k] Navigate  [Enter] Play  [/] Search  [p] Playlists  [r] Refresh  [q] Quit"))

	return sb.String()
}

// libraryColWidths returns column widths [title, artist, album] based on terminal width.
func libraryColWidths(w int) [3]int {
	avail := w - 4 - 8 // subtract # col and time col
	if avail < 30 {
		avail = 30
	}
	t := avail / 3
	return [3]int{t, t, t}
}

// truncate shortens a string to max length with ellipsis.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 4 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
