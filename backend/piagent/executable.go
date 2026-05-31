package piagent

import "path/filepath"

func piExecutableCandidates(homeDir string) []string {
	return []string{
		filepath.Join(homeDir, "AppData", "Local", "pnpm", "pi.CMD"),
		filepath.Join(homeDir, "AppData", "Local", "pnpm", "pi.cmd"),
		filepath.Join(homeDir, "AppData", "Local", "pnpm", "pi"),
		filepath.Join(homeDir, "AppData", "Roaming", "npm", "pi.cmd"),
		filepath.Join(homeDir, "AppData", "Roaming", "npm", "pi.CMD"),
		filepath.Join(homeDir, "AppData", "Roaming", "npm", "pi"),
	}
}
