// Package ui — app_client.go implements the root BubbleTea model for the gtmpc
// client-server TUI. It manages screen transitions, health checks, and shared state
// (API client + audio engine). This is separate from the existing local-mode app.go.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/audio"
	"github.com/jscyril/golang_music_player/internal/ui/screens"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
	"github.com/jscyril/golang_music_player/pkg/stats"
)

// ClientScreen enumerates the active screen in client mode.
type ClientScreen int

const (
	ClientScreenLogin ClientScreen = iota
	ClientScreenRegister
	ClientScreenLibrary
	ClientScreenPlayer
	ClientScreenPlaylist
	ClientScreenStats
	ClientScreenError
)

// ClientApp is the root BubbleTea model for the client-server TUI.
type ClientApp struct {
	client   *apiclient.APIClient
	engine   *audio.AudioEngine
	username string

	screen ClientScreen

	// Sub-screens
	loginScreen    screens.LoginScreen
	registerScreen screens.RegisterScreen
	libraryScreen  screens.LibraryScreen
	playerScreen   screens.PlayerScreen
	playlistScreen screens.PlaylistScreen
	statsScreen    screens.StatsScreen

	// Playback state
	allTracks []apiclient.Track
	queueIdx  int

	// Session statistics tracker
	stats *stats.Stats

	// Error state
	connErr string

	width  int
	height int
}

// healthCheckMsg carries the result of the startup health check.
type healthCheckMsg struct {
	ok  bool
	err error
}

// NewClientApp creates the root application model.
func NewClientApp(client *apiclient.APIClient, engine *audio.AudioEngine, width, height int) ClientApp {
	s := stats.New()
	return ClientApp{
		client:      client,
		engine:      engine,
		screen:      ClientScreenLogin,
		loginScreen: screens.NewLoginScreen(client, width, height),
		stats:       s,
		width:       width,
		height:      height,
	}
}

// doHealthCheck checks if the backend is reachable.
func (a ClientApp) doHealthCheck() tea.Cmd {
	return func() tea.Msg {
		_, err := a.client.CheckHealth()
		return healthCheckMsg{ok: err == nil, err: err}
	}
}

// Init starts the health check and login screen.
func (a ClientApp) Init() tea.Cmd {
	return tea.Batch(
		a.doHealthCheck(),
		a.loginScreen.Init(),
	)
}

// Update handles all messages and routes them to the active screen.
func (a ClientApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to all sub-screens
		a.loginScreen = screens.NewLoginScreen(a.client, a.width, a.height)
		a.registerScreen = screens.NewRegisterScreen(a.client, a.width, a.height)

	case tea.KeyMsg:
		// Global: ctrl+c always quits
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		// 'q' quits from library/stats/error screens (player handles its own q)
		if a.screen == ClientScreenLibrary || a.screen == ClientScreenStats || a.screen == ClientScreenError {
			if msg.String() == "q" {
				return a, tea.Quit
			}
		}
		// [4] navigates to Stats from Library
		if a.screen == ClientScreenLibrary && msg.String() == "4" {
			a.statsScreen = screens.NewStatsScreen(a.stats, a.width, a.height)
			a.screen = ClientScreenStats
			return a, a.statsScreen.Init()
		}
		// '?' shows help in any screen (we just render it inline)
		if msg.String() == "?" {
			// No-op for now — help is rendered in each screen's view
		}

	case healthCheckMsg:
		if !msg.ok {
			a.connErr = fmt.Sprintf("Cannot connect to server: %v\n\nPress [r] to retry or [q] to quit.", msg.err)
			a.screen = ClientScreenError
		}
		return a, nil

	// ── Screen transition messages ────────────────────────────────────────
	case screens.AuthSuccessMsg:
		a.username = msg.Username
		a.client.SetToken(msg.Token)
		a.libraryScreen = screens.NewLibraryScreen(a.client, a.width, a.height)
		a.screen = ClientScreenLibrary
		return a, a.libraryScreen.Init()

	case screens.GoToRegisterMsg:
		a.registerScreen = screens.NewRegisterScreen(a.client, a.width, a.height)
		a.screen = ClientScreenRegister
		return a, a.registerScreen.Init()

	case screens.GoToLoginMsg:
		a.loginScreen = screens.NewLoginScreen(a.client, a.width, a.height)
		a.screen = ClientScreenLogin
		return a, a.loginScreen.Init()

	case screens.PlayTrackMsg:
		a.allTracks = msg.AllTracks
		a.queueIdx = msg.Index
		// Record the play event in session stats
		a.stats.RecordPlay(msg.Track.ID, msg.Track.Title, msg.Track.Artist, msg.Track.Album, msg.Track.DurationSeconds)
		// Start playback via HTTP streaming
		streamURL := a.client.StreamURL(msg.Track.ID)
		if err := playHTTPTrack(a.engine, streamURL, a.client.Token); err != nil {
			// Non-fatal: stay on library but show error in player
		}
		a.playerScreen = screens.NewPlayerScreen(a.engine, a.client, msg.Track, msg.AllTracks, msg.Index, a.width, a.height)
		a.screen = ClientScreenPlayer
		return a, a.playerScreen.Init()

	case screens.NextTrackMsg:
		a.queueIdx++
		if a.queueIdx >= len(a.allTracks) {
			a.queueIdx = 0
		}
		if len(a.allTracks) > 0 {
			track := a.allTracks[a.queueIdx]
			a.stats.RecordPlay(track.ID, track.Title, track.Artist, track.Album, track.DurationSeconds)
			streamURL := a.client.StreamURL(track.ID)
			playHTTPTrack(a.engine, streamURL, a.client.Token)
			a.playerScreen = screens.NewPlayerScreen(a.engine, a.client, track, a.allTracks, a.queueIdx, a.width, a.height)
		}
		return a, a.playerScreen.Init()

	case screens.PrevTrackMsg:
		if a.queueIdx > 0 {
			a.queueIdx--
			track := a.allTracks[a.queueIdx]
			a.stats.RecordPlay(track.ID, track.Title, track.Artist, track.Album, track.DurationSeconds)
			streamURL := a.client.StreamURL(track.ID)
			playHTTPTrack(a.engine, streamURL, a.client.Token)
			a.playerScreen = screens.NewPlayerScreen(a.engine, a.client, track, a.allTracks, a.queueIdx, a.width, a.height)
			return a, a.playerScreen.Init()
		}
		// Already first track — go back to library
		a.screen = ClientScreenLibrary
		return a, nil

	case screens.BackToLibraryMsg:
		a.screen = ClientScreenLibrary
		return a, nil

	case screens.GoToStatsMsg:
		a.statsScreen = screens.NewStatsScreen(a.stats, a.width, a.height)
		a.screen = ClientScreenStats
		return a, a.statsScreen.Init()

	case screens.GoToPlaylistMsg:
		a.playlistScreen = screens.NewPlaylistScreen(a.client, a.width, a.height)
		a.screen = ClientScreenPlaylist
		return a, a.playlistScreen.Init()
	}

	// ── Delegate to active screen ─────────────────────────────────────────
	var cmd tea.Cmd
	switch a.screen {
	case ClientScreenLogin:
		a.loginScreen, cmd = a.loginScreen.Update(msg)
	case ClientScreenRegister:
		a.registerScreen, cmd = a.registerScreen.Update(msg)
	case ClientScreenLibrary:
		a.libraryScreen, cmd = a.libraryScreen.Update(msg)
	case ClientScreenPlayer:
		a.playerScreen, cmd = a.playerScreen.Update(msg)
	case ClientScreenPlaylist:
		a.playlistScreen, cmd = a.playlistScreen.Update(msg)
	case ClientScreenStats:
		a.statsScreen, cmd = a.statsScreen.Update(msg)
	case ClientScreenError:
		if kMsg, ok := msg.(tea.KeyMsg); ok {
			switch kMsg.String() {
			case "r":
				a.connErr = ""
				a.screen = ClientScreenLogin
				return a, a.doHealthCheck()
			case "q":
				return a, tea.Quit
			}
		}
	}

	return a, cmd
}

// View renders the currently active screen.
func (a ClientApp) View() string {
	// Status bar at bottom
	var content string

	switch a.screen {
	case ClientScreenLogin:
		content = a.loginScreen.View()
	case ClientScreenRegister:
		content = a.registerScreen.View()
	case ClientScreenLibrary:
		content = a.renderWithStatusBar(a.libraryScreen.View())
	case ClientScreenPlayer:
		content = a.playerScreen.View()
	case ClientScreenPlaylist:
		content = a.playlistScreen.View()
	case ClientScreenStats:
		content = a.renderWithStatusBar(a.statsScreen.View())
	case ClientScreenError:
		content = a.renderErrorScreen()
	}

	return content
}

// renderWithStatusBar wraps content with the bottom status bar.
func (a ClientApp) renderWithStatusBar(content string) string {
	state := a.engine.GetState()
	var nowPlaying string
	if state.CurrentTrack != nil {
		statusIcon := "▶"
		if state.Status == 2 { // paused
			statusIcon = "⏸"
		}
		nowPlaying = fmt.Sprintf("%s %s — %s", statusIcon, state.CurrentTrack.Title, state.CurrentTrack.Artist)
	} else {
		nowPlaying = "No track playing"
	}

	statusBar := styles.StatusBarStyle.
		Width(a.width).
		Render(fmt.Sprintf("  %s  |  User: %s  |  [4] Stats  [q] Quit", nowPlaying, a.username))

	// Trim content to fit height
	lines := strings.Split(content, "\n")
	maxLines := a.height - 1
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	trimmed := strings.Join(lines, "\n")

	return trimmed + "\n" + statusBar
}

// renderErrorScreen renders the server unreachable error screen.
func (a ClientApp) renderErrorScreen() string {
	msg := styles.ErrorStyle.Render("⚠  Server Unreachable") + "\n\n" +
		styles.SubtitleStyle.Render(a.connErr) + "\n\n" +
		styles.HelpStyle.Render("[r] Retry  [q] Quit")

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorError).
		Padding(2, 4).
		Render(msg)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, card)
}

// playHTTPTrack starts audio playback from an HTTP stream URL.
// It uses the engine's HTTP streaming support.
func playHTTPTrack(engine *audio.AudioEngine, streamURL, token string) error {
	return engine.PlayFromURL(streamURL, token)
}

// RunClientApp starts the BubbleTea program in client mode.
func RunClientApp(client *apiclient.APIClient, engine *audio.AudioEngine) error {
	app := NewClientApp(client, engine, 80, 24)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
