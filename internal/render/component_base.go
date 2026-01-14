package render

type Viewable interface {
	View() string
}

type ComponentBase struct {
	v          Viewable
	updateChan chan ComponentUpdate
	onMount    []func()
}

func NewComponentBase(v Viewable) ComponentBase {
	return ComponentBase{v: v}
}

func (b *ComponentBase) OnMount(f func()) {
	b.onMount = append(b.onMount, f)
}

// SetUpdateChan implements [Component].
func (b *ComponentBase) SetUpdateChan(c chan ComponentUpdate) {
	b.updateChan = c

	for _, f := range b.onMount {
		f()
	}
}

func (b *ComponentBase) Render() {
	b.updateChan <- ComponentUpdate{NewText: b.v.View()}
}

func (b *ComponentBase) RenderWait() {
	text := b.v.View()

	// Send 2 updates, the second one to check that the first has gone through
	b.updateChan <- ComponentUpdate{NewText: text, Wait: true}
	b.updateChan <- ComponentUpdate{NewText: text, Wait: false, NoRender: true}
}

func (b *ComponentBase) End() {
	b.RenderWait()
	close(b.updateChan)
}
