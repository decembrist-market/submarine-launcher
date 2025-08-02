package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runExecution(path string) error {
	cmd := exec.Command(path, "-launcher")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Start()
}

// TryRunGame пытается запустить игру
func TryRunGame(dataDir string) {
	fmt.Print("Запуск игры...\n")

	// Получаем имя исполняемого файла для текущей платформы
	gameFile := GetExecutableForPlatform()
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
