package kit

import (
	"path/filepath"

	"github.com/PondWader/kit/include"
	"github.com/PondWader/kit/pkg/lang"
)

func (b *installBinding) loadMod(modName string) (*lang.Environment, error) {
	// Read embeded library
	modCode, err := include.Lib.Open(filepath.Join(modName, modName+".kit"))
	if err != nil {
		return nil, err
	}
	defer modCode.Close()

	env := lang.NewEnv()
	b.Load(env)

	err = env.ExecuteReader(modCode)
	if err != nil {
		return nil, err
	}
	return env, nil
}
