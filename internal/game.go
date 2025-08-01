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
			ShowStyledMessage(Error, "Ошибка при запуске игры: "+err.Error())
			return
		}
	} else {
		ShowStyledMessage(Error, "Файл игры не найден: "+gameFilePath)
		return
	}
	fmt.Println("Игра запущена")
	os.Exit(0)
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
