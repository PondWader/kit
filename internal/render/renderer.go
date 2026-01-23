package render

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/PondWader/kit/internal/ansi"
	"golang.org/x/term"
)

type ComponentUpdate struct {
	NewText  string
	Wait     bool
	NoRender bool
}

type Component interface {
	Bind(chan ComponentUpdate, *MountedComponent)
	View() string
}

type Term struct {
	Out *os.File
	In  *os.File

	components []*MountedComponent
	updateChan chan chan struct{}

	lastLineCount int
}

func NewTerm(in, out *os.File) *Term {
	ch := make(chan chan struct{}, 1)

	t := &Term{
		Out:        out,
		components: make([]*MountedComponent, 0),
		updateChan: ch,
	}

	go func() {
		for cb := range ch {
			t.render()
			if cb != nil {
				cb <- struct{}{}
			}
		}
	}()

	go t.inputReader(in)

	return t
}

func (r *Term) Println(v ...any) {
	r.Mount(staticComponent{fmt.Sprint(v...)})
}

func (r *Term) Update() {
	// If there is already a pending update, no need to send twice
	select {
	case r.updateChan <- nil:
	default:
	}
}

func (r *Term) UpdateAndWait() {
	cb := make(chan struct{})
	r.updateChan <- cb
	<-cb
}

func (r *Term) render() {
	var sb strings.Builder

	// Save cursor position if last component is receiving input
	var hasInput bool
	if len(r.components) > 0 {
		last := r.components[len(r.components)-1]
		if last.input != nil && last.displayed {
			sb.WriteString("\u001B7")
			hasInput = true
		}
	}

	// Clear lines
	for range r.lastLineCount - 1 {
		sb.WriteString("\u001B[1A\r\x1b[K")
	}
	resetLen := sb.Len()

	for _, c := range r.components {
		if c.Text == "" {
			sb.WriteRune('\r')
			continue
		}
		fmt.Fprint(&sb, c.Text)
		c.displayed = true
	}

	// Restore cursor position
	if hasInput {
		sb.WriteString("\u001B8")
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

func (t *Term) inputReader(in *os.File) {
	var b [1024]byte
	for {
		n, err := in.Read(b[:])
		if err != nil {
			return
		}

		rcv := t.components[len(t.components)-1]
		if rcv.input == nil {
			continue
		}

		str := string(b[:n])
		// Go up a line to cancel out new line
		if str[len(str)-1] == '\n' {
			t.Out.WriteString("\u001B[1A")
		}
		rcv.input <- str
	}
}

func (t *Term) Mount(c Component) {
	mc := &MountedComponent{
		Component: c,
	}
	t.components = append(t.components, mc)

	ch := make(chan ComponentUpdate)
	c.Bind(ch, mc)

	mc.Text = c.View()
	t.Update()

	go func() {
		for update := range ch {
			mc.Text = update.NewText
			if update.NoRender {
				continue
			}
			if update.Wait {
				t.UpdateAndWait()
			} else {
				t.Update()
			}
		}
	}()
}

func (t *Term) Stop() {
	close(t.updateChan)
}

func lineWidth(line string) int {
	width := 0
	inAnsi := false

	for _, c := range line {
		if inAnsi || ansi.IsMarker(c) {
			inAnsi = !ansi.IsTerminator(c)
		} else {
			width += utf8.RuneLen(c)
		}
	}

	return width
}

func countLines(line string, width int) int {
	count := 0
	for line := range strings.SplitSeq(line, "\n") {
		count++

		lineWidth := lineWidth(line)
		if lineWidth > width {
			count += int(float64(lineWidth) / float64(width))
		}
	}

	return count
}

type staticComponent struct {
	Text string
}

// SetUpdateChan implements [Component].
func (s staticComponent) Bind(c chan ComponentUpdate, _ *MountedComponent) {
	close(c)
}

// View implements [Component].
func (s staticComponent) View() string {
	return s.Text
}

var _ Component = (*staticComponent)(nil)

type MountedComponent struct {
	Text      string
	Component Component
	input     chan<- string
	displayed bool
}

func (mc *MountedComponent) Input() <-chan string {
	rcv := make(chan string)
	mc.input = rcv
	return rcv
}
