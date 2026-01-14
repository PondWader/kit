package render

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

type ComponentUpdate struct {
	NewText  string
	Wait     bool
	NoRender bool
}

type Component interface {
	SetUpdateChan(chan ComponentUpdate)
	View() string
}

type MountedComponent struct {
	Text      string
	Component Component
}

type Renderer struct {
	Out        *os.File
	components []*MountedComponent
	updateChan chan chan struct{}

	lastLineCount int
}

func NewRenderer(out *os.File) *Renderer {
	ch := make(chan chan struct{}, 1)

	r := &Renderer{
		Out:        out,
		components: make([]*MountedComponent, 0),
		updateChan: ch,
	}

	go func() {
		for cb := range ch {
			r.render()
			if cb != nil {
				cb <- struct{}{}
			}
		}
	}()

	return r
}

func (r *Renderer) Println(v ...any) {
	r.Mount(staticComponent{fmt.Sprint(v...)})
}

func (r *Renderer) Update() {
	// If there is already a pending update, no need to send twice
	select {
	case r.updateChan <- nil:
	default:
	}
}

func (r *Renderer) UpdateAndWait() {
	cb := make(chan struct{})
	r.updateChan <- cb
	<-cb
}

func (r *Renderer) render() {
	var sb strings.Builder
	for range r.lastLineCount - 1 {
		sb.WriteString("\u001B[1A\r\x1b[K")
	}
	resetLen := sb.Len()

	for _, c := range r.components {
		if c.Text == "" {
			sb.WriteRune('\r')
			continue
		}
		fmt.Fprintln(&sb, c.Text)
	}
	str := sb.String()

	terminalWidth := 80
	if term.IsTerminal(0) {
		if width, _, err := term.GetSize(0); err == nil {
			terminalWidth = width
		}
	}
	lineCount := countLines(str[resetLen:], terminalWidth)
	r.lastLineCount = lineCount

	os.Stdout.WriteString(str)
}

func (r *Renderer) Mount(c Component) {
	ch := make(chan ComponentUpdate)
	c.SetUpdateChan(ch)

	mc := &MountedComponent{
		Text:      c.View(),
		Component: c,
	}
	r.components = append(r.components, mc)
	r.Update()

	go func() {
		for update := range ch {
			mc.Text = update.NewText
			if update.NoRender {
				continue
			}
			if update.Wait {
				r.UpdateAndWait()
			} else {
				r.Update()
			}
		}
	}()
}

func (r *Renderer) Stop() {
	close(r.updateChan)
}

func isAnsiMarker(r rune) bool {
	return r == '\x1b'
}

func isAnsiTerminator(r rune) bool {
	return (r >= 0x40 && r <= 0x5a) || (r == 0x5e) || (r >= 0x60 && r <= 0x7e)
}

func lineWidth(line string) int {
	width := 0
	ansi := false

	for _, r := range line {
		if ansi || isAnsiMarker(r) {
			ansi = !isAnsiTerminator(r)
		} else {
			width += utf8.RuneLen(r)
		}
	}

	return width
}

func countLines(line string, width int) int {
	lineCount := 0
	for line := range strings.SplitSeq(line, "\n") {
		lineCount++

		lineWidth := lineWidth(line)
		if lineWidth > width {
			lineCount += int(float64(lineWidth) / float64(width))
		}
	}

	return lineCount
}

type staticComponent struct {
	Text string
}

// SetUpdateChan implements [Component].
func (s staticComponent) SetUpdateChan(c chan ComponentUpdate) {
	close(c)
}

// View implements [Component].
func (s staticComponent) View() string {
	return s.Text
}

var _ Component = (*staticComponent)(nil)
