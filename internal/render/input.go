package render

import (
	"strings"
)

type TextInput struct {
	ComponentBase
	prompt string
	input  strings.Builder
}

var _ Component = (*TextInput)(nil)

func NewTextInput(prompt string) (i *TextInput) {
	i = &TextInput{
		prompt: prompt,
	}
	i.ComponentBase = NewComponentBase(i)
	i.OnMount(func() {
		go i.inputHandler()
	})
	return i
}

func (i *TextInput) inputHandler() {
	for in := range i.Input() {
		i.input.WriteString(in)
		i.Render()
	}
}

func (i *TextInput) View() string {
	return i.prompt + i.input.String()
}
