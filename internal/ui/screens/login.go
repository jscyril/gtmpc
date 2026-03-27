// Package screens provides individual TUI screens for the gtmpc client.
// login.go implements the login screen with username/password inputs.
package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
)

// AuthSuccessMsg is sent when login succeeds, carrying the token and username.
type AuthSuccessMsg struct {
	Token    string
	Username string
}

// GoToRegisterMsg is sent when the user presses 'r' to switch to the register screen.
type GoToRegisterMsg struct{}

// LoginScreen implements the login form.
type LoginScreen struct {
	client    *apiclient.APIClient
	inputs    []textinput.Model
	focused   int
	err       string
	loading   bool
	width     int
	height    int
}

const (
	loginFieldUser = 0
	loginFieldPass = 1
)

// NewLoginScreen creates a new LoginScreen.
func NewLoginScreen(client *apiclient.APIClient, width, height int) LoginScreen {
	userInput := textinput.New()
	userInput.Placeholder = "Username"
	userInput.Focus()
	userInput.Width = 36

	passInput := textinput.New()
	passInput.Placeholder = "Password"
	passInput.EchoMode = textinput.EchoPassword
	passInput.EchoCharacter = '•'
	passInput.Width = 36

	return LoginScreen{
		client:  client,
		inputs:  []textinput.Model{userInput, passInput},
		focused: 0,
		width:   width,
		height:  height,
	}
}

// loginCmd performs the actual login API call.
type loginResultMsg struct {
	resp *apiclient.LoginResponse
	err  error
}

func (s LoginScreen) doLogin() tea.Cmd {
	return func() tea.Msg {
		resp, err := s.client.Login(apiclient.LoginRequest{
			Username: s.inputs[loginFieldUser].Value(),
			Password: s.inputs[loginFieldPass].Value(),
		})
		return loginResultMsg{resp: resp, err: err}
	}
}

// Init initializes the screen.
func (s LoginScreen) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the login screen.
func (s LoginScreen) Update(msg tea.Msg) (LoginScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case loginResultMsg:
		s.loading = false
		if msg.err != nil {
			s.err = humanizeError(msg.err)
		} else {
			return s, func() tea.Msg {
				return AuthSuccessMsg{Token: msg.resp.Token, Username: msg.resp.Username}
			}
		}

	case tea.KeyMsg:
		if s.loading {
			return s, nil
		}
		s.err = "" // clear error on any keypress

		switch msg.String() {
		case "tab", "down":
			s.focused = (s.focused + 1) % len(s.inputs)
			for i := range s.inputs {
				if i == s.focused {
					s.inputs[i].Focus()
				} else {
					s.inputs[i].Blur()
				}
			}
		case "shift+tab", "up":
			s.focused = (s.focused - 1 + len(s.inputs)) % len(s.inputs)
			for i := range s.inputs {
				if i == s.focused {
					s.inputs[i].Focus()
				} else {
					s.inputs[i].Blur()
				}
			}
		case "enter":
			if s.inputs[loginFieldUser].Value() == "" || s.inputs[loginFieldPass].Value() == "" {
				s.err = "Please enter both username and password"
				return s, nil
			}
			s.loading = true
			return s, s.doLogin()
		case "r":
			return s, func() tea.Msg { return GoToRegisterMsg{} }
		}
	}

	// Update focused input
	var cmds []tea.Cmd
	for i := range s.inputs {
		var cmd tea.Cmd
		s.inputs[i], cmd = s.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return s, tea.Batch(cmds...)
}

// View renders the login screen.
func (s LoginScreen) View() string {
	var sb strings.Builder

	sb.WriteString(styles.TitleStyle.Render("🎵 gtmpc — Login") + "\n\n")
	sb.WriteString(styles.SubtitleStyle.Render("Enter your credentials to continue") + "\n\n")

	sb.WriteString("Username\n")
	sb.WriteString(s.inputs[loginFieldUser].View() + "\n\n")

	sb.WriteString("Password\n")
	sb.WriteString(s.inputs[loginFieldPass].View() + "\n\n")

	if s.loading {
		sb.WriteString(styles.SubtitleStyle.Render("Logging in...") + "\n")
	} else if s.err != "" {
		sb.WriteString(styles.ErrorStyle.Render("✗ "+s.err) + "\n")
	} else {
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.HelpStyle.Render("[Tab] Switch field  [Enter] Login  [r] Register"))

	card := styles.CardStyle.Render(sb.String())

	// Center the card on screen
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, card)
}

// humanizeError converts API errors to user-friendly messages.
func humanizeError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "unauthorized") || strings.Contains(msg, "invalid credentials"):
		return "Invalid username or password"
	case strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host"):
		return "Cannot connect to server"
	case strings.Contains(msg, "timeout"):
		return "Request timed out — server may be slow"
	case strings.Contains(msg, "conflict") || strings.Contains(msg, "already exists"):
		return "Username already taken"
	default:
		return "Error: " + msg
	}
}
