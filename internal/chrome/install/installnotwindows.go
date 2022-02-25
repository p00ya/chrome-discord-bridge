//go:build darwin || linux

package install

import (
	"os/user"
	"path/filepath"
)

// CurrentUser installs the manifest for the calling user.
func CurrentUser(m Manifest) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	return User(m, usr.HomeDir)
}

// User creates and installs a Chrome manifest to a user-specific
// directory.
func User(m Manifest, homeDir string) error {
	buf, err := m.Marshal()
	if err != nil {
		return err
	}
	name := filepath.Join(homeDir, userSubDir, m.Filename())
	return install(name, buf)
}

// System creates and installs a Chrome manifest to the system-wide
// directory.
func System(m Manifest) error {
	buf, err := m.Marshal()
	if err != nil {
		return err
	}
	name := filepath.Join(systemDir, m.Filename())
	return install(name, buf)
}
