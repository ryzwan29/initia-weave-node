package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
)

type ClickableItem struct {
	displayText map[bool]string
	clicked     bool
	handleFn    func() error
	x           int
	y           int
	w           map[bool]int
}

func NewClickableItem(displayText map[bool]string, handleFn func() error) *ClickableItem {
	w := map[bool]int{
		true:  len(displayText[true]),
		false: len(displayText[false]),
	}
	return &ClickableItem{
		displayText: displayText,
		clicked:     false,
		handleFn:    handleFn,
		x:           0,
		y:           0,
		w:           w,
	}
}

func (ci *ClickableItem) Update(msg tea.Msg) error {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft &&
			msg.X >= ci.x && msg.X < ci.x+ci.w[ci.clicked] &&
			msg.Y == ci.y {
			ci.clicked = !ci.clicked
			if ci.handleFn != nil {
				return ci.handleFn()
			}
		}
	}
	return nil
}

func (ci *ClickableItem) View() string {
	return ci.displayText[ci.clicked]
}

func (ci *ClickableItem) UpdatePosition(cleanText string) error {
	lines := strings.Split(cleanText, "\n")
	for y, line := range lines {
		if strings.Contains(line, ci.displayText[ci.clicked]) {
			ci.y = y
			ci.x = strings.Index(line, ci.displayText[ci.clicked])
			return nil
		}
	}
	return fmt.Errorf("text '%s' not found in rendered lines", ci.displayText[ci.clicked])
}

type Clickable struct {
	Items []*ClickableItem
}

func NewClickable(items ...*ClickableItem) *Clickable {
	return &Clickable{Items: items}
}

func (c *Clickable) Init() tea.Cmd {
	return tea.EnableMouseCellMotion
}

func (c *Clickable) ClickableUpdate(msg tea.Msg) error {
	for _, item := range c.Items {
		if err := item.Update(msg); err != nil {
			return err
		}
	}
	return nil
}

func (c *Clickable) ClickableView(index int) string {
	item := c.Items[index]
	return item.displayText[item.clicked]
}

func (c *Clickable) PostUpdate() tea.Cmd {
	return tea.DisableMouse
}

func (c *Clickable) ClickableUpdatePositions(text string) error {
	cleanText := styles.StripANSI(text)
	lines := strings.Split(cleanText, "\n")

	remainingItems := make(map[string][]*ClickableItem)
	for _, item := range c.Items {
		display := item.displayText[item.clicked]
		remainingItems[display] = append(remainingItems[display], item)
	}

	for y, line := range lines {
		for display, items := range remainingItems {
			startIndex := 0
			for {
				idx := strings.Index(line[startIndex:], display)
				if idx == -1 {
					break
				}
				if len(items) > 0 {
					item := items[0]
					item.x = startIndex + idx
					item.y = y

					remainingItems[display] = items[1:]
					items = remainingItems[display]
				}
				startIndex += idx + len(display)
			}
		}
	}

	for display, items := range remainingItems {
		if len(items) > 0 {
			return fmt.Errorf("text '%s' not found for all occurrences in rendered lines", display)
		}
	}

	return nil
}
