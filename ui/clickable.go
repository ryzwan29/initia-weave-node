package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
)

type Clickable struct {
	displayText map[bool]string
	clicked     bool
	x           int
	y           int
	w           map[bool]int
}

func NewClickable(displayText map[bool]string) *Clickable {
	w := map[bool]int{
		true:  len(displayText[true]),
		false: len(displayText[false]),
	}
	return &Clickable{
		displayText: displayText,
		clicked:     false,
		x:           0,
		y:           0,
		w:           w,
	}
}

func (c *Clickable) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

func (c *Clickable) ClickableUpdate(msg tea.Msg, handleFn func() error) error {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft &&
			msg.X >= c.x && msg.X < c.x+c.w[c.clicked] &&
			msg.Y == c.y {
			c.clicked = !c.clicked
			if err := handleFn(); err != nil {
				return fmt.Errorf("error executing click handler: %w", err)
			}
		}
	}

	return nil
}

func (c *Clickable) ClickableView() string {
	return c.displayText[c.clicked]
}

func (c *Clickable) PostUpdate() tea.Cmd {
	return tea.DisableMouse
}

func (c *Clickable) ClickableUpdatePosition(text string) error {
	cleanText := styles.StripANSI(text)

	lines := strings.Split(cleanText, "\n")
	for y, line := range lines {
		if strings.Contains(line, c.displayText[c.clicked]) {
			c.y = y
			c.x = strings.Index(line, c.displayText[c.clicked])
			return nil
		}
	}
	return fmt.Errorf("text '%s' not found in rendered lines", c.displayText[c.clicked])
}
