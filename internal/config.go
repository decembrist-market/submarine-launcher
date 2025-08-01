package internal

type GameExecutables struct {
	Windows string
	Linux   string
	Darwin  string // macOS
}

// Глобальная конфигурация
var (
	GameFolderName   = "SubmarineGame"
	RemoteVersionURL = "https://static.decembrist.org/submarine-game/version.yaml"
	VersionFileName  = "version.yaml"
	GameFileName     = "submarine"
	ArchiveURL       = "https://static.decembrist.org/submarine-game/submarine.zip"
	HashURL          = "https://static.decembrist.org/submarine-game/submarine.zip.sha256"

	GameExes = GameExecutables{
		Windows: "submarine.exe", // Исполняемый файл для Windows
		Linux:   "submarine",     // Для Linux
		Darwin:  "submarine",     // Для macOS
	}
)
