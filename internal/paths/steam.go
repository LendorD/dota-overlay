package paths

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func FindDotaLogPath() (string, error) {
	var steamPath string
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		k, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Valve\Steam`, registry.QUERY_VALUE)
	}
	if err == nil {
		steamPath, _, _ = k.GetStringValue("InstallPath")
		k.Close()
	}

	libs := []string{filepath.Join(steamPath, "steamapps")}
	vdf := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")
	if f, err := os.Open(vdf); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), `"path"`) {
				p := strings.Split(scanner.Text(), "\"")[3]
				libs = append(libs, filepath.Join(p, "steamapps"))
			}
		}
		f.Close()
	}

	for _, lib := range libs {
		path := filepath.Join(lib, "common", "dota 2 beta", "game", "dota", "console.log")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}
