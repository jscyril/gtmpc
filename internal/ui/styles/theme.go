// Package styles defines the Lip Gloss color palette and reusable UI styles
// for the gtmpc TUI client. All screens and components should import from here
// to ensure visual consistency.
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette matching the React web frontend theme
const (
	ColorPrimary    = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary  = lipgloss.Color("#10B981") // Green
	ColorAccent     = lipgloss.Color("#F59E0B") // Amber
	ColorBackground = lipgloss.Color("#1F2937") // Dark gray
	ColorText       = lipgloss.Color("#F9FAFB") // Near white
	ColorMuted      = lipgloss.Color("#6B7280") // Gray
	ColorError      = lipgloss.Color("#EF4444") // Red
	ColorSuccess    = lipgloss.Color("#10B981") // Green
	ColorSurface    = lipgloss.Color("#374151") // Slightly lighter dark
	ColorBorder     = lipgloss.Color("#4B5563") // Border gray
)

// CardStyle is used for login/register centered cards
var CardStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(ColorPrimary).
	Padding(1, 2).
	Width(50)

// TitleStyle is for screen titles and headings
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPrimary)

// SubtitleStyle is for secondary labels
var SubtitleStyle = lipgloss.NewStyle().
	Foreground(ColorMuted)

// InputStyle is for text input fields
var InputStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(ColorBorder).
	Padding(0, 1)

// FocusedInputStyle is used when a text input has keyboard focus
var FocusedInputStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(ColorPrimary).
	Padding(0, 1)

// TableHeaderStyle is used for the header row of track tables
var TableHeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorMuted).
	BorderBottom(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(ColorBorder)

// SelectedRowStyle highlights the currently selected table row
var SelectedRowStyle = lipgloss.NewStyle().
	Background(ColorPrimary).
	Foreground(ColorText).
	Bold(true)

// ActiveRowStyle highlights the currently playing track
var ActiveRowStyle = lipgloss.NewStyle().
	Foreground(ColorSecondary).
	Bold(true)

// ProgressBarStyle is for the playback progress bar
var ProgressBarStyle = lipgloss.NewStyle().
	Foreground(ColorPrimary)

// ProgressBarEmptyStyle is for the unfilled portion of the progress bar
var ProgressBarEmptyStyle = lipgloss.NewStyle().
	Foreground(ColorBorder)

// StatusBarStyle is for the bottom status bar
var StatusBarStyle = lipgloss.NewStyle().
	Background(ColorSurface).
	Foreground(ColorText).
	Padding(0, 1)

// ErrorStyle renders error messages
var ErrorStyle = lipgloss.NewStyle().
	Foreground(ColorError).
	Bold(true)

// SuccessStyle renders success messages
var SuccessStyle = lipgloss.NewStyle().
	Foreground(ColorSuccess).
	Bold(true)

// HelpStyle renders keybind hints
var HelpStyle = lipgloss.NewStyle().
	Foreground(ColorMuted)

// NowPlayingStyle is for the now-playing track info
var NowPlayingStyle = lipgloss.NewStyle().
	Foreground(ColorSecondary).
	Bold(true)

// CenteredStyle centers content horizontally
func CenteredStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
}
