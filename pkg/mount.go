package kit

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/PondWader/kit/pkg/db"
)

type Mount struct {
	k       *Kit
	i       *db.Installation
	actions []db.MountAction
}

type MountOptions struct {
	Name    string
	Repo    string
	Version string
}

func NewMount(k *Kit, o MountOptions) (*Mount, error) {
	i, err := k.DB.BeginInstallation(o.Name, o.Repo, o.Version, false)
	if err != nil {
		return nil, err
	}
	return &Mount{k, i, nil}, nil
}

func LoadMount(k *Kit, id int64) (*Mount, error) {
	actions, err := k.DB.GetInstallMountActions(id)
	if err != nil {
		return nil, err
	}
	return &Mount{k: k, actions: actions}, nil
}

func (m *Mount) recordAction(action string, data map[string]string) error {
	actionRecord, err := m.i.RecordMountAction(action, data)
	if err != nil {
		return err
	}
	m.actions = append(m.actions, actionRecord)
	return nil
}

func (m *Mount) LinkBin(target, linkName string) error {
	return m.recordAction("link_bin", map[string]string{
		"target":   target,
		"linkName": linkName,
	})
}

func (m *Mount) Enable(dir string) error {
	// TODO: maybe track these actions in the DB before performing them to rollback if the activation does not complete
	for _, a := range m.actions {
		switch a.Action {
		case "link_bin":
			linkPath := filepath.Join(m.k.Home.BinDir(), a.Data["linkName"])
			if err := m.k.Home.Remove(linkPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			target := filepath.Join(dir, a.Data["target"])
			relTarget, err := filepath.Rel(filepath.Dir(linkPath), target)
			if err != nil {
				return err
			}
			if err := m.k.Home.Symlink(relTarget, linkPath); err != nil {
				return err
			}
		default:
			return errors.New("unknown action \"" + a.Action + "\"")
		}
	}

	if m.i != nil {
		if err := m.i.SetActive(true); err != nil {
			return err
		}
		return m.i.Commit()
	}
	return nil
}

func (m *Mount) Close() error {
	return m.i.Rollback()
}
