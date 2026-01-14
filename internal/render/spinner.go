package render

import (
	"sync"
	"time"

	"github.com/PondWader/kit/internal/ansi"
)

type Spinner struct {
	ComponentBase

	mu           sync.Mutex
	Frames       []string
	text         string
	currentFrame int
	updateChan   chan ComponentUpdate
	ticker       *time.Ticker

	stopped bool
	success bool
}

func NewSpinner(text string) *Spinner {
	s := &Spinner{
		Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		text:   text,
	}
	s.ComponentBase = NewComponentBase(s)

	s.OnMount(func() {
		ticker := time.NewTicker(time.Millisecond * 80)
		s.ticker = ticker

		go func() {
			for range ticker.C {
				s.mu.Lock()
				s.currentFrame += 1
				if s.currentFrame >= len(s.Frames) {
					s.currentFrame = 0
				}
				s.mu.Unlock()

				s.Render()
			}
		}()
	})

	return s
}

var _ Component = (*Spinner)(nil)

// View implements [Component].
func (s *Spinner) View() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		if s.success {
			return ansi.Green("✔") + " " + s.text
		}
		return ""
	}
	return ansi.Cyan(s.Frames[s.currentFrame]) + " " + s.text
}

func (s *Spinner) Stop() {
	s.ticker.Stop()

	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	s.mu.Unlock()

	s.End()
}

func (s *Spinner) Succeed(msg string) {
	s.mu.Lock()
	s.text = msg
	s.success = true
	s.mu.Unlock()
	s.Stop()
}
