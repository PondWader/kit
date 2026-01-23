package render

type Viewable interface {
	View() string
}

type ComponentBase struct {
	v          Viewable
	updateChan chan ComponentUpdate
	onMount    []func()
	mount      *MountedComponent
}

func NewComponentBase(v Viewable) ComponentBase {
	return ComponentBase{v: v}
}

func (b *ComponentBase) OnMount(f func()) {
	b.onMount = append(b.onMount, f)
}

// Bind implements [Component].
func (b *ComponentBase) Bind(c chan ComponentUpdate, mount *MountedComponent) {
	b.updateChan = c
	b.mount = mount

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

func (b *ComponentBase) Input() <-chan string {
	return b.mount.Input()
}
