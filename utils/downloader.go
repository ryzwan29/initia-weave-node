package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/initia-labs/weave/styles"
)

type Downloader struct {
	progress progress.Model
	total    int64
	current  int64
	text     string
	url      string
	dest     string
	done     bool
}

func NewDownloader(text, url, dest string) *Downloader {
	return &Downloader{
		progress: progress.New(progress.WithGradient(string(styles.Cyan), "#008B8B")),
		total:    0,
		text:     text,
		url:      url,
		dest:     dest,
	}
}

func (m *Downloader) startDownload() tea.Cmd {
	return func() tea.Msg {
		// Use a high buffer size for faster I/O operations
		const bufferSize = 8192 // 8 KB buffer

		resp, err := http.Get(m.url)
		if err != nil {
			m.done = true
			return nil
		}
		defer resp.Body.Close()

		m.total = resp.ContentLength
		if m.total <= 0 {
			m.total = 1
		}

		file, err := os.Create(m.dest)
		if err != nil {
			m.done = true
			return nil
		}
		defer file.Close()

		buffer := make([]byte, bufferSize)
		var totalDownloaded int64
		for {
			n, err := resp.Body.Read(buffer)
			if err != nil && err != io.EOF {
				m.done = true
				return nil
			}
			if n == 0 {
				break
			}

			// Write to file and update progress in a single step
			if _, err := file.Write(buffer[:n]); err != nil {
				m.done = true
				return nil
			}

			totalDownloaded += int64(n)
			m.current = totalDownloaded

			// Simulate network delay for smoother updates
			time.Sleep(10 * time.Millisecond)
		}

		m.done = true
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
			return m, tea.Quit
		}
	}

	model, cmd := m.progress.Update(msg)
	m.progress = model.(progress.Model)
	return m, cmd
}

func (m *Downloader) View() string {
	if m.done {
		return fmt.Sprintf("%sDownload Complete!\nTotal Size: %d bytes\n", styles.CorrectMark, m.total)
	}
	percentage := float64(m.current) / float64(m.total)
	return fmt.Sprintf("%s: %s / %s \n%s", m.text, ByteCountSI(m.current), ByteCountSI(m.total), m.progress.ViewAs(percentage))
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
