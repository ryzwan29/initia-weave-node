package utils

import (
	"bytes"
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
	err      error
}

func NewDownloader(text, url, dest string) *Downloader {
	return &Downloader{
		progress: progress.New(progress.WithGradient(string(styles.Cyan), string(styles.DarkCyan))),
		total:    0,
		text:     text,
		url:      url,
		dest:     dest,
		err:      nil,
	}
}

func (m *Downloader) GetError() error {
	return m.err
}

func (m *Downloader) startDownload() tea.Cmd {
	return func() tea.Msg {
		const bufferSize = 65536

		// Send a GET request to the URL
		resp, err := http.Get(m.url)
		if err != nil {
			m.SetError(fmt.Errorf("failed to connect to URL: %w", err))
			return nil
		}
		defer resp.Body.Close()

		// Check if the response status is not 200 OK
		if resp.StatusCode != http.StatusOK {
			m.SetError(fmt.Errorf("failed to download: received status code %d", resp.StatusCode))
			return nil
		}

		// Get the total size of the file
		m.total = resp.ContentLength
		if m.total <= 0 {
			// If Content-Length is not provided, we set a default value for safety
			m.total = 1
		}

		// Create the destination file
		file, err := os.Create(m.dest)
		if err != nil {
			m.SetError(fmt.Errorf("failed to create destination file: %w", err))
			return nil
		}
		defer file.Close()

		// Prepare to download the file in chunks
		buffer := make([]byte, bufferSize)
		var totalDownloaded int64
		for {
			n, err := resp.Body.Read(buffer)
			if err != nil && err != io.EOF {
				m.SetError(fmt.Errorf("error during file download: %w", err))
				return nil
			}
			if n == 0 {
				break
			}

			// Write the downloaded chunk to the file
			if _, err := file.Write(buffer[:n]); err != nil {
				m.SetError(fmt.Errorf("failed to write to file: %w", err))
				return nil
			}

			// Update the progress
			totalDownloaded += int64(n)
			m.current = totalDownloaded
		}

		// Now, check if the file is a valid .tar.lz4 file
		if err := m.validateTarLz4Header(); err != nil {
			m.SetError(fmt.Errorf("invalid file format: %w", err))
			return nil
		}

		// Set completion when the download finishes successfully
		m.SetCompletion(true)
		return nil
	}
}

// LZ4 magic number for LZ4 frame format
var lz4MagicNumber = []byte{0x04, 0x22, 0x4D, 0x18}

// validateTarLz4Header checks if the downloaded file is a valid .tar.lz4 file based on the file header.
func (m *Downloader) validateTarLz4Header() error {
	// Open the .lz4 file
	file, err := os.Open(m.dest)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes of the file (header)
	header := make([]byte, 4)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	// Check if the header matches the LZ4 magic number
	if !bytes.Equal(header, lz4MagicNumber) {
		return fmt.Errorf("invalid file format: the file is not a valid .lz4 file")
	}

	// If the header matches, we assume it's a valid .lz4 file.
	// You could continue checking the contents further if needed, but this verifies the file is compressed with LZ4.

	return nil
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
	if m.err != nil {
		return ""
	}

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

// SetCompletion allows you to manually set the completion state for testing purposes.
func (m *Downloader) SetCompletion(complete bool) {
	m.done = complete
}

// SetError allows you to manually set an error for testing purposes.
func (m *Downloader) SetError(err error) {
	m.err = err
	m.done = true // Mark as done when an error occurs
}
