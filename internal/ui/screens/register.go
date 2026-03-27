// Package screens — register.go implements the account registration screen.
package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jscyril/golang_music_player/internal/ui/styles"
	"github.com/jscyril/golang_music_player/pkg/apiclient"
)

// GoToLoginMsg is sent when the user presses 'l' to switch back to the login screen.
type GoToLoginMsg struct{}

// RegisterSuccessMsg is sent after successful registration.
type RegisterSuccessMsg struct {
	Username string
}

// RegisterScreen implements the user registration form.
type RegisterScreen struct {
	client  *apiclient.APIClient
	inputs  []textinput.Model
	focused int
	err     string
	success string
	loading bool
	width   int
	height  int
}

const (
	regFieldUser    = 0
	regFieldPass    = 1
	regFieldConfirm = 2
)

// NewRegisterScreen creates a new RegisterScreen.
func NewRegisterScreen(client *apiclient.APIClient, width, height int) RegisterScreen {
	userInput := textinput.New()
	userInput.Placeholder = "Username"
	userInput.Focus()
	userInput.Width = 36

	passInput := textinput.New()
	passInput.Placeholder = "Password (min 6 chars)"
	passInput.EchoMode = textinput.EchoPassword
	passInput.EchoCharacter = '•'
	passInput.Width = 36

	confirmInput := textinput.New()
	confirmInput.Placeholder = "Confirm Password"
	confirmInput.EchoMode = textinput.EchoPassword
	confirmInput.EchoCharacter = '•'
	confirmInput.Width = 36

	return RegisterScreen{
		client:  client,
		inputs:  []textinput.Model{userInput, passInput, confirmInput},
		focused: 0,
		width:   width,
		height:  height,
	}
}

type registerResultMsg struct {
	username string
	err      error
}

func (s RegisterScreen) doRegister() tea.Cmd {
	return func() tea.Msg {
		resp, err := s.client.Register(apiclient.RegisterRequest{
			Username: s.inputs[regFieldUser].Value(),
			Password: s.inputs[regFieldPass].Value(),
		})
		if err != nil {
			return registerResultMsg{err: err}
		}
		return registerResultMsg{username: resp.UserID}
	}
}

// Init initializes the screen.
func (s RegisterScreen) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages for the register screen.
func (s RegisterScreen) Update(msg tea.Msg) (RegisterScreen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case registerResultMsg:
		s.loading = false
		if msg.err != nil {
			s.err = humanizeError(msg.err)
		} else {
			s.success = "Account created! Redirecting to login..."
			return s, func() tea.Msg {
				return GoToLoginMsg{}
			}
		}

	case tea.KeyMsg:
		if s.loading {
			return s, nil
		}
		s.err = ""

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
			user := s.inputs[regFieldUser].Value()
			pass := s.inputs[regFieldPass].Value()
			confirm := s.inputs[regFieldConfirm].Value()

			if user == "" || pass == "" || confirm == "" {
				s.err = "All fields are required"
				return s, nil
			}
			if len(pass) < 6 {
				s.err = "Password must be at least 6 characters"
				return s, nil
			}
			if pass != confirm {
				s.err = "Passwords do not match"
				return s, nil
			}
			s.loading = true
			return s, s.doRegister()
		case "l", "esc":
			return s, func() tea.Msg { return GoToLoginMsg{} }
		}
	}

	var cmds []tea.Cmd
	for i := range s.inputs {
		var cmd tea.Cmd
		s.inputs[i], cmd = s.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return s, tea.Batch(cmds...)
}

// View renders the register screen.
func (s RegisterScreen) View() string {
	var sb strings.Builder

	sb.WriteString(styles.TitleStyle.Render("🎵 gtmpc — Create Account") + "\n\n")

	sb.WriteString("Username\n")
	sb.WriteString(s.inputs[regFieldUser].View() + "\n\n")

	sb.WriteString("Password\n")
	sb.WriteString(s.inputs[regFieldPass].View() + "\n\n")

	sb.WriteString("Confirm Password\n")
	sb.WriteString(s.inputs[regFieldConfirm].View() + "\n\n")

	if s.loading {
		sb.WriteString(styles.SubtitleStyle.Render("Creating account...") + "\n")
	} else if s.err != "" {
		sb.WriteString(styles.ErrorStyle.Render("✗ "+s.err) + "\n")
	} else if s.success != "" {
		sb.WriteString(styles.SuccessStyle.Render("✓ "+s.success) + "\n")
	} else {
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.HelpStyle.Render("[Tab] Switch field  [Enter] Create Account  [l/Esc] Back to Login"))

	card := styles.CardStyle.Render(sb.String())
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, card)
}
