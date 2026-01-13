package lang

import "os"

func InitRuntime(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	_ = f
	return nil
}

type Runtime struct{}
