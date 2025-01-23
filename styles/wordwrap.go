package styles

import (
	"github.com/muesli/reflow/wordwrap"
)

var (
	defaultBreakpoints = []rune{'-', ','}
	defaultNewline     = []rune{'\n'}
)

func NewWordwrapWriter(limit int) *wordwrap.WordWrap {
	return &wordwrap.WordWrap{
		Limit:        limit,
		Breakpoints:  defaultBreakpoints,
		Newline:      defaultNewline,
		KeepNewlines: true,
	}
}

func WordwrapBytes(b []byte, limit int) []byte {
	f := NewWordwrapWriter(limit)
	_, _ = f.Write(b)
	_ = f.Close()

	return f.Bytes()
}

func WordwrapString(s string, limit int) string {
	return string(WordwrapBytes([]byte(s), limit))
}
