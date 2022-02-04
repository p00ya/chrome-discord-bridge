package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Manifest models the Chrome native messaging host manifest JSON.
//
// See the official Chrome documentation at:
// https://developer.chrome.com/docs/apps/nativeMessaging/#native-messaging-host
type Manifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	AllowedOrigins []string `json:"allowed_origins"`
	Typ            string   `json:"type"`
}

// manifestType is the (only supported) value for the "type" field in the
// manifest.
const manifestType = "stdio"

// Marshal returns on-disk encoding of the manifest.
func (m Manifest) Marshal() ([]byte, error) {
	m.Typ = manifestType
	return json.Marshal(m)
}

// Filename is the approriate name for the manifest file (with no path).
func (m Manifest) Filename() string {
	return m.Name + ".json"
}

// InstallCurrentUser installs the manifest for the calling user.
func InstallCurrentUser(m Manifest) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	return InstallUser(m, usr.HomeDir)
}

// InstallUser creates and installs a Chrome manifest to a user-specific
// directory on macOS.
func InstallUser(m Manifest, homeDir string) error {
	buf, err := m.Marshal()
	if err != nil {
		return err
	}
	name := filepath.Join(homeDir, userSubDir, m.Filename())
	return install(name, buf)
}

// InstallSystem creates and install a Chrome manifest to the system-wide
// directory on macOS.
func InstallSystem(m Manifest) error {
	buf, err := m.Marshal()
	if err != nil {
		return err
	}
	name := filepath.Join(systemDir, m.Filename())
	return install(name, buf)
}

// install writes the serialized manifest buffer to the given path.
func install(name string, buf []byte) error {
	if err := os.WriteFile(name, buf, 0644); err != nil {
		return fmt.Errorf(`writing manifest: %w`, err)
	}
	return nil
}
