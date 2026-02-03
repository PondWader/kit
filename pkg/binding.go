package kit

import (
	"runtime"

	"github.com/PondWader/kit/pkg/lang"
	"github.com/PondWader/kit/pkg/lang/values"
)

type installBinding struct{}

func (b *installBinding) CreateSys() *values.Object {
	o := values.NewObject()
	o.Put("OS", values.Of(runtime.GOOS))
	o.Put("ARCH", values.Of(runtime.GOARCH))
	return o
}

func (b *installBinding) Load(env *lang.Environment) {
	env.Set("sys", b.CreateSys().Val())
}
