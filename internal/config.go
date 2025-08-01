package internal

type GameExecutables struct {
	Windows string
	Linux   string
	Darwin  string
}

var (
	LauncherVersion  = "0.0.0"
	GameFolderName   = "SubmarineGame"
	RemoteVersionURL = "https://static.decembrist.org/submarine-game/version.yaml"
	LauncherURL      = "https://static.decembrist.org/submarine-game/SubmarineLauncher.exe"
	VersionFileName  = "version.yaml"
	ArchiveURL       = "https://static.decembrist.org/submarine-game/submarine.zip"
	HashURL          = "https://static.decembrist.org/submarine-game/submarine.zip.sha256"

	GameExes = GameExecutables{
		Windows: "submarine.exe",
		Linux:   "submarine",
		Darwin:  "submarine",
	}
)
