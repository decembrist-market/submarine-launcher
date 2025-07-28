package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

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

func TryRunGame(dataDir string) {
	fmt.Print("Запуск игры...\n")
	gameFilePath := filepath.Join(dataDir, getExecutableName(GameFileName))
	if _, err := os.Stat(gameFilePath); err == nil {
		err := runExecution(gameFilePath)
		if err != nil {
			ShowExitMessage(Error, "Ошибка при запуске игры:", err)
			return
		}
	} else {
		ShowExitMessage(Error, "Файл игры не найден:", fmt.Errorf("путь: %s", gameFilePath))
		return
	}
	fmt.Println("Игра запущена")
	os.Exit(0)
}

func GetGameDirection(launcherPath string) (string, error) {
	launcherDirPath := filepath.Dir(launcherPath)
	var gameDirPath string
	if filepath.Base(launcherDirPath) == GameFolderName {
		gameDirPath = launcherDirPath
	} else {
		gameDirPath = filepath.Join(launcherDirPath, GameFolderName)
		if _, err := os.Stat(gameDirPath); os.IsNotExist(err) {
			err := os.Mkdir(gameDirPath, 0755)
			if err != nil {
				return "", err
			}
			fmt.Println("Папка игры создана.")
		}
	}
	return gameDirPath, nil
}
