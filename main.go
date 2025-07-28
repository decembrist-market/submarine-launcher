package main

import (
	"fmt"
	"os"
	"path/filepath"
	"submarine-launcher/internal"
)

func main() {
	fmt.Println("🚢 Запуск лаунчера Submarine...")
	launcherPath, err := os.Executable()
	if err != nil {
		internal.ShowExitMessage(internal.Error, "Ошибка при получении пути к исполняемому файлу:", err)
		return
	}

	gameDirPath, err := internal.GetGameDirection(launcherPath)
	if err != nil {
		internal.ShowExitMessage(internal.Error, "Ошибка при получении директории игры:", err)
		return
	}

	localVersionPath := filepath.Join(gameDirPath, internal.VersionFileName)
	if _, err := os.Stat(localVersionPath); os.IsNotExist(err) {
		fmt.Println("Игра не найдена, установить игру? (y/n)")
		isPlayerAgree := internal.CheckAnswer()
		if isPlayerAgree {
			internal.TryUnzipGame(gameDirPath, launcherPath)
		}
		internal.ShowExitMessage(internal.Info)
		return
	} else if err != nil {
		internal.ShowExitMessage(internal.Error, "Ошибка при проверке версии игры:", err)
		return
	}

	localVersion, err := os.ReadFile(localVersionPath)
	if err != nil {
		internal.ShowExitMessage(internal.Error, "Ошибка при чтении файла с версией:", err)
		return
	}
	remoteVersion, err := internal.GetRemoteVersion()
	if err != nil {
		internal.ShowExitMessage(internal.Error, "Ошибка при проверке удалённой версии:", err)
		return
	}

	if string(localVersion) == (remoteVersion) {
		internal.TryRunGame(gameDirPath)
		return
	} else {
		fmt.Println("Обновить игру перед запуском? (y/n)")
		isPlayerAgree := internal.CheckAnswer()
		if isPlayerAgree {
			internal.TryUnzipGame(gameDirPath, launcherPath)
			fmt.Println("Игра обновлена до версии", remoteVersion)
		}
		internal.TryRunGame(gameDirPath)
		return
	}
}
