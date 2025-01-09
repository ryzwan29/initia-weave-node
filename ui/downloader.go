package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/analytics"
	"github.com/initia-labs/weave/client"
	"github.com/initia-labs/weave/styles"
)

type Downloader struct {
	progress   progress.Model
	total      int64
	current    int64
	text       string
	url        string
	dest       string
	done       bool
	err        error
	validateFn func(string) error
}

func NewDownloader(text, url, dest string, validateFn func(string) error) *Downloader {
	return &Downloader{
		progress:   progress.New(progress.WithGradient(string(styles.Cyan), string(styles.DarkCyan))),
		total:      0,
		text:       text,
		url:        url,
		dest:       dest,
		err:        nil,
		validateFn: validateFn,
	}
}

func (m *Downloader) GetError() error {
	return m.err
}

func (m *Downloader) startDownload() tea.Cmd {
	return func() tea.Msg {
		httpClient := client.NewHTTPClient()
		if err := httpClient.DownloadAndValidateFile(m.url, m.dest, &m.current, &m.total, m.validateFn); err != nil {
			m.SetError(err)
			return nil
		}

		// Set completion when the download finishes successfully
		m.SetCompletion(true)
		return nil
	}
}

func (m *Downloader) Init() tea.Cmd {
	return tea.Batch(m.tick(), m.startDownload())
}

func (m *Downloader) Update(msg tea.Msg) (*Downloader, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		return m, m.tick()

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			analytics.TrackEvent(analytics.Interrupted, analytics.NewEmptyEvent())
			return m, tea.Quit
		}
	}

	model, cmd := m.progress.Update(msg)
	m.progress = model.(progress.Model)
	return m, cmd
}

func (m *Downloader) View() string {
	if m.err != nil {
		return ""
	}

	if m.done {
		return fmt.Sprintf("%sDownload Complete!\nTotal Size: %d bytes\n", styles.CorrectMark, m.total)
	}
	percentage := float64(m.current) / float64(m.total)
	return fmt.Sprintf("\n %s: %s / %s \n %s", m.text, ByteCountSI(m.current), ByteCountSI(m.total), m.progress.ViewAs(percentage))
}

func (m *Downloader) GetCompletion() bool {
	return m.done
}

func (m *Downloader) tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.3f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// SetCompletion allows you to manually set the completion state for testing purposes.
func (m *Downloader) SetCompletion(complete bool) {
	m.done = complete
}

// SetError allows you to manually set an error for testing purposes.
func (m *Downloader) SetError(err error) {
	m.err = err
	m.done = true // Mark as done when an error occurs
}
