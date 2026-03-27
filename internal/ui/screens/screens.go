// Package screens — screens.go defines shared message types exchanged between
// screens and the root ClientApp model (app_client.go).
package screens

// GoToStatsMsg is sent when the user navigates to the Stats screen.
type GoToStatsMsg struct{}

// ToggleLikeMsg is sent when the user toggles a track like.
type ToggleLikeMsg struct {
	TrackID string
	Title   string
}
