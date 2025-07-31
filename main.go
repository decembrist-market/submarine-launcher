package main

import (
	"fmt"
	"os"
	"path/filepath"
	"submarine-launcher/internal"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Включаем поддержку цветов в Windows терминале
	if err := internal.EnableWindowsColors(); err == nil {
		// Очистка экрана для красивого отображения
		fmt.Print("\033[2J\033[H")
	}

	launcherPath, err := os.Executable()
	if err != nil {
		internal.ShowStyledMessage("error", "Ошибка при получении пути к исполняемому файлу: "+err.Error())
		internal.ShowExitMessage(internal.Error, "")
		return
	}

	gameDirPath, err := internal.GetGameDirection(launcherPath)
	if err != nil {
		internal.ShowStyledMessage("error", "Ошибка при получении директории игры: "+err.Error())
		internal.ShowExitMessage(internal.Error, "")
		return
	}

	// Проверяем наличие игры
	localVersionPath := filepath.Join(gameDirPath, internal.VersionFileName)
	gameInstalled := true
	needsUpdate := false

	if _, err := os.Stat(localVersionPath); os.IsNotExist(err) {
		gameInstalled = false
	} else if err != nil {
		internal.ShowStyledMessage("error", "Ошибка при проверке версии игры: "+err.Error())
		internal.ShowExitMessage(internal.Error, "")
		return
	}

	// Если игра установлена, проверяем обновления
	if gameInstalled {
		localVersion, err := os.ReadFile(localVersionPath)
		if err != nil {
			internal.ShowStyledMessage("error", "Ошибка при чтении файла с версией: "+err.Error())
			internal.ShowExitMessage(internal.Error, "")
			return
		}

		remoteVersion, err := internal.GetRemoteVersion()
		if err != nil {
			internal.ShowStyledMessage("warning", "Не удалось проверить обновления, запускаем игру...")
		} else {
			needsUpdate = string(localVersion) != remoteVersion
		}
	}

	// Создаем и запускаем TUI модель
	model := internal.NewTUIModel(gameInstalled, needsUpdate)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	finalModel, err := p.Run()
	if err != nil {
		internal.ShowStyledMessage("error", "Ошибка интерфейса: "+err.Error())
		return
	}

	// Обрабатываем выбор пользователя
	tuiModel := finalModel.(internal.TUIModel)
	if !tuiModel.WasSelected() {
		return
	}
	choice := tuiModel.GetChoice()

	if !gameInstalled {
		// Игра не установлена
		switch choice {
		case 0: // Установить игру
			internal.ShowStyledMessage("info", "Начинается установка игры...")
			internal.TryUnzipGame(gameDirPath, launcherPath)
			internal.ShowStyledMessage("success", "Игра успешно установлена!")
		case 1: // Выход
			internal.ShowStyledMessage("info", "До свидания! 👋")
			return
		}
	} else if needsUpdate {
		// Игра установлена, но нужно обновление
		switch choice {
		case 0: // Обновить игру
			internal.ShowStyledMessage("info", "Начинается обновление игры...")
			internal.TryUnzipGame(gameDirPath, launcherPath)
			internal.ShowStyledMessage("success", "Игра успешно обновлена!")
			internal.TryRunGame(gameDirPath)
		case 1: // Запустить игру
			internal.TryRunGame(gameDirPath)
		case 2: // Выход
			internal.ShowStyledMessage("info", "До свидания! 👋")
			return
		}
	} else {
		// Игра установлена и актуальна
		switch choice {
		case 0: // Запустить игру
			internal.TryRunGame(gameDirPath)
		case 1: // Выход
			internal.ShowStyledMessage("info", "До свидания! 👋")
			return
		}
	}

	internal.ShowExitMessage(internal.Info, "")
}
