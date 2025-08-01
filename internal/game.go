package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetExecutableForPlatform возвращает имя исполняемого файла для текущей платформы
func GetExecutableForPlatform(gameExes GameExecutables) string {
	switch runtime.GOOS {
	case "windows":
		return gameExes.Windows
	case "linux":
		return gameExes.Linux
	case "darwin":
		return gameExes.Darwin
	default:
		return gameExes.Linux // По умолчанию используем Linux версию
	}
}

// getExecutableName - устаревшая функция, оставлена для совместимости
func getExecutableName(baseName string) string {
	if runtime.GOOS == "windows" {
		return baseName + ".exe"
	}
	return baseName
}
func runExecution(path string) error {
	cmd := exec.Command(path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Start()
}

// TryRunGame пытается запустить игру
func TryRunGame(dataDir string) {
	fmt.Print("Запуск игры...\n")

	// Получаем имя исполняемого файла для текущей платформы
	gameFile := GetExecutableForPlatform(GameExes)
	gamePath := filepath.Join(dataDir, gameFile)

	// Проверяем существование файла и запускаем игру
	if _, err := os.Stat(gamePath); err == nil {
		err := runExecution(gamePath)
		if err != nil {
			ShowStyledMessage(Error, "Ошибка при запуске игры: "+err.Error())
			return
		}
		fmt.Printf("Игра запущена: %s\n", filepath.Base(gamePath))
		os.Exit(0)
		return
	}

	// Если файл не найден, показываем ошибку
	ShowStyledMessage(Error, fmt.Sprintf("Исполняемый файл игры не найден: %s", gamePath))
}

func GetGameDirPath(launcherPath string) string {
	launcherDirPath := filepath.Dir(launcherPath)
	var gameDirPath string
	if filepath.Base(launcherDirPath) == GameFolderName {
		gameDirPath = launcherDirPath
	} else {
		gameDirPath = filepath.Join(launcherDirPath, GameFolderName)
	}
	return gameDirPath
}
