package styles

import "github.com/muesli/reflow/wordwrap"

const (
	DefaultPadding int = 2
)

// Wrapper is an interface that represents a common WrapView function.
type Wrapper interface {
	WrapView(contentLen int) string
}

type BaseWrapper struct {
	ContentWidth int
}

func (b *BaseWrapper) WrapView(content string) string {
	return wordwrap.String(content, b.ContentWidth-DefaultPadding)
}
