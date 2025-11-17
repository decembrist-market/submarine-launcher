package internal

import "runtime"

type GameExecutables struct {
	Windows string
	Linux   string
	Darwin  string
}

type TtsExecutables struct {
	Windows string
	Linux   string
	Darwin  string
}

type DownloadURLs struct {
	Windows string
	Linux   string
	Darwin  string
}

type DownloadGameURLs struct {
	Windows     string
	Linux       string
	DarwinArm64 string
	DarwinIntel string
}

var (
	LauncherVersion     = "0.0.12"
	GameFolderName      = "SubmarineGame"
	RemoteManifestURL   = "https://static.decembrist.org/submarine-game/launcher-manifest.yaml"
	GameVersionFileName = "version.yaml"

	LauncherURLs = DownloadURLs{
		Windows: "https://static.decembrist.org/submarine-game/windows/SubmarineLauncher.exe",
		Linux:   "https://static.decembrist.org/submarine-game/linux/SubmarineLauncher",
		Darwin:  "https://static.decembrist.org/submarine-game/macos/SubmarineLauncher",
	}

	ArchiveURLs = DownloadGameURLs{
		Windows:     "https://static.decembrist.org/submarine-game/windows/submarine.zip",
		Linux:       "https://static.decembrist.org/submarine-game/linux/submarine.zip",
		DarwinArm64: "https://static.decembrist.org/submarine-game/macos-arm64/submarine.zip",
		DarwinIntel: "https://static.decembrist.org/submarine-game/macos-intel/submarine.zip",
	}

	HashURLs = DownloadURLs{
		Windows: "https://static.decembrist.org/submarine-game/windows/submarine.zip.sha256",
		Linux:   "https://static.decembrist.org/submarine-game/linux/submarine.zip.sha256",
		Darwin:  "https://static.decembrist.org/submarine-game/macos/submarine.zip.sha256",
	}

	GameExes = GameExecutables{
		Windows: "submarine.exe",
		Linux:   "submarine.x86_64",
		Darwin:  "Submarine Godot.app/Contents/MacOS/Submarine Godot",
	}

	TtsExes = TtsExecutables{
		Windows: "tts.exe",
		Linux:   "tts",
		Darwin:  "tts",
	}
)

func GetLauncherURL() string {
	switch runtime.GOOS {
	case "windows":
		return LauncherURLs.Windows
	case "linux":
		return LauncherURLs.Linux
	case "darwin":
		return LauncherURLs.Darwin
	default:
		return LauncherURLs.Windows // fallback
	}
}

func GetArchiveURL() string {
	switch runtime.GOOS {
	case "windows":
		return ArchiveURLs.Windows
	case "linux":
		return ArchiveURLs.Linux
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return ArchiveURLs.DarwinArm64
		}
		return ArchiveURLs.DarwinIntel
	default:
		return ArchiveURLs.Windows // fallback
	}
}

func GetHashURL() string {
	switch runtime.GOOS {
	case "windows":
		return HashURLs.Windows
	case "linux":
		return HashURLs.Linux
	case "darwin":
		return HashURLs.Darwin
	default:
		return HashURLs.Windows // fallback
	}
}

func GetExecutableForPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return GameExes.Windows
	case "linux":
		return GameExes.Linux
	case "darwin":
		return GameExes.Darwin
	default:
		return GameExes.Windows
	}
}

func GetTtsExecutableForPlatform() string {
	switch runtime.GOOS {
	case "windows":
		return TtsExes.Windows
	case "linux":
		return TtsExes.Linux
	case "darwin":
		return TtsExes.Darwin
	default:
		return TtsExes.Windows
	}
}
