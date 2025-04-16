package fetch

import (
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding  = 2
	maxWidth = 80
)

type ProgressReader struct {
	io.Reader
	Total      int64
	ReadSoFar  int64
	LastUpdate time.Time
	IsSpinner  bool
	percent    float64
	progress   progress.Model
}

type tickMsg time.Time

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if err != nil {
		return n, err
	}

	pr.ReadSoFar += int64(n)
	pr.percent = float64(pr.ReadSoFar) / float64(pr.Total)

	return n, nil
}

func newProgress(r io.Reader, total int64) *ProgressReader {
	pr := &ProgressReader{
		Reader:   r,
		Total:    total,
		progress: progress.New(progress.WithDefaultGradient()),
	}

	return pr
}

func (pr *ProgressReader) run() error {
	_, err := tea.NewProgram(pr).Run()
	if err != nil {
		return err
	}

	return nil
}

func (pr *ProgressReader) Init() tea.Cmd {
	return tickCmd()
}

func (pr *ProgressReader) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return pr, tea.Quit

	case tea.WindowSizeMsg:
		pr.progress.Width = msg.Width - padding*2 - 4
		if pr.progress.Width > maxWidth {
			pr.progress.Width = maxWidth
		}
		return pr, nil

	case tickMsg:
		// TODO: integrate spinner

		if pr.progress.Percent() == 1.0 {
			return pr, tea.Quit
		}

		cmd := pr.progress.SetPercent(pr.percent)

		// Note that you can also use progress.Model.SetPercent to set the
		// percentage value explicitly, too.
		// cmd := pr.progress.IncrPercent(0.25)
		return pr, tea.Batch(tickCmd(), cmd)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := pr.progress.Update(msg)
		pr.progress = progressModel.(progress.Model)
		return pr, cmd

	default:
		return pr, nil
	}
}

func (pr *ProgressReader) View() string {
	pad := strings.Repeat(" ", padding)
	return pad + pr.progress.View()
}

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
