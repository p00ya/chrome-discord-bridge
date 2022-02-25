package install

import (
	"fmt"
	"os"
	"path/filepath"
)

import "golang.org/x/sys/windows/registry"

// CurrentUser writes a Chrome manifest in the current directory, and registers
// it in the Windows registry under HKEY_CURRENT_USER.
func CurrentUser(m Manifest) error {
	return writeManifestAndRegister(m, registry.CURRENT_USER)
}

// System writes a Chrome manifest in the current directory, and registers
// it in the Windows registry under HKEY_LOCAL_MACHINE.
func System(m Manifest) error {
	return writeManifestAndRegister(m, registry.LOCAL_MACHINE)
}

func writeManifestAndRegister(m Manifest, root registry.Key) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	manifestPath, err := writeManifest(cwd, m)
	if err != nil {
		return err
	}
	return register(root, m.Name, manifestPath)
}

func writeManifest(dir string, m Manifest) (string, error) {
	buf, err := m.Marshal()
	if err != nil {
		return "", err
	}

	manifestPath := filepath.Join(dir, m.Filename())
	if err = install(manifestPath, buf); err != nil {
		return "", err
	}

	return manifestPath, nil
}

// keyPath is the path under the registry root to register Chrome native
// messaging hosts.
const keyPath = `SOFTWARE\Google\Chrome\NativeMessagingHosts`

// register registers the native messaging host in the Windows registry.
func register(root registry.Key, name string, manifestPath string) error {
	p := fmt.Sprintf(`%s\%s`, keyPath, name)
	k, _, err := registry.CreateKey(registry.CURRENT_USER, p, registry.CREATE_SUB_KEY|registry.SET_VALUE)
	if err != nil {
		return err
	}
	return k.SetStringValue("", manifestPath)
}
