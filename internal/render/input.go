package render

import (
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/term"
)

type TextInput struct {
	ComponentBase
	prompt   string
	input    strings.Builder
	readC    chan string
	complete bool
	oldState *term.State
}

var _ Component = (*TextInput)(nil)

func NewTextInput(prompt string, secret bool) (i *TextInput) {
	i = &TextInput{
		prompt: prompt,
		readC:  make(chan string),
	}
	i.ComponentBase = NewComponentBase(i)
	i.OnMount(func() {
		if secret {
			i.oldState, _ = disableEcho(int(os.Stdin.Fd()))
		}

		go i.inputHandler()
	})
	return i
}

func (i *TextInput) inputHandler() {
	for in := range i.Input() {
		if in[len(in)-1] == '\n' {
			i.input.WriteString(in[:len(in)-1])
			i.complete = true
			if i.oldState != nil {
				term.Restore(int(os.Stdin.Fd()), i.oldState)
			}
			i.End()
			i.readC <- i.input.String()
			break
		}

		i.input.WriteString(in)
		i.Render()
	}
}

func (i *TextInput) Read() string {
	return <-i.readC
}

func (i *TextInput) View() string {
	if i.complete {
		return ""
	}
	return i.prompt + i.input.String()
}

// DisableEcho disables terminal echo for the input file descriptor.
// Returns the original terminal state that can be passed to RestoreEcho.
func disableEcho(fd int) (*term.State, error) {
	oldState, err := term.GetState(fd)
	if err != nil {
		return nil, err
	}

	// Get the current termios
	var termios syscall.Termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&termios))); errno != 0 {
		return nil, errno
	}

	// Disable echo
	termios.Lflag &^= syscall.ECHO

	// Set the new termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), syscall.TCSETS, uintptr(unsafe.Pointer(&termios))); errno != 0 {
		return nil, errno
	}

	return oldState, nil
}
