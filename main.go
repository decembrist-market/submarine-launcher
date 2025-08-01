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
		internal.ShowExitMessage(internal.Error, "Ошибка при получении пути к исполняемому файлу: "+err.Error())
		return
	}

	gameDirPath := internal.GetGameDirPath(launcherPath)

	// Основной цикл лаунчера
	for {
		// Проверяем наличие игры
		localVersionPath := filepath.Join(gameDirPath, internal.VersionFileName)
		gameDirExist := true
		gameInstalled := true
		needsUpdate := false

		if _, err := os.Stat(gameDirPath); os.IsNotExist(err) {
			gameDirExist = false
			gameInstalled = false
		} else if err != nil {
			internal.ShowExitMessage(internal.Error, "Ошибка при проверке папки игры: ")
			return
		}

		if gameDirExist {
			if _, err := os.Stat(localVersionPath); os.IsNotExist(err) {
				gameInstalled = false
			} else if err != nil {
				internal.ShowExitMessage(internal.Error, "Ошибка при проверке версии игры: "+err.Error())
				return
			}
		}

		// Если игра установлена, проверяем обновления
		if gameInstalled {
			localVersion, err := os.ReadFile(localVersionPath)
			if err != nil {
				internal.ShowExitMessage(internal.Error, "Ошибка при чтении файла с версией: "+err.Error())
				return
			}

			remoteVersion, err := internal.GetRemoteVersion()
			if err != nil {
				internal.ShowStyledMessage(internal.Warn, "Не удалось проверить обновления, запускаем игру...")
			} else {
				needsUpdate = string(localVersion) != remoteVersion
			}
		}

		// Создаем и запускаем TUI модель
		model := internal.NewTUIModel(gameInstalled, needsUpdate)
		p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

		finalModel, err := p.Run()
		if err != nil {
			internal.ShowStyledMessage(internal.Error, "Ошибка интерфейса: "+err.Error())
			return
		}

		// Обрабатываем выбор пользователя
		tuiModel := finalModel.(internal.TUIModel)
		if !tuiModel.WasSelected() {
			return
		}
		choice := tuiModel.GetChoice()

		shouldExit := false

		if !gameInstalled {
			// Игра не установлена
			switch choice {
			case 0: // Установить игру
				if !gameDirExist {
					err := os.Mkdir(gameDirPath, 0755)
					if err != nil {
						// Показываем ошибку и возвращаемся в меню
						fmt.Printf("Ошибка создания директории: %s\nНажмите Enter для продолжения...\n", err.Error())
						fmt.Scanln()
						continue
					}
				}

				// Выходим из TUI режима для установки
				fmt.Print("\033[2J\033[H") // Очищаем экран
				fmt.Println("Начинается установка игры...")
				err = internal.TryUnzipGame(gameDirPath, launcherPath)
				if err != nil {
					// Показываем ошибку и возвращаемся в меню
					fmt.Printf("Ошибка установки: %s\nНажмите Enter для продолжения...\n", err.Error())
					fmt.Scanln()
					continue
				}

				fmt.Println("\n✅ Игра успешно установлена!")
				fmt.Println("Теперь вы можете запустить игру из меню.")
				fmt.Print("Нажмите Enter для возврата в меню...")
				fmt.Scanln()
				// Продолжаем цикл, чтобы показать обновленное меню
				continue
			case 1: // Выход
				shouldExit = true
			}
		} else if needsUpdate {
			// Игра установлена, но нужно обновление
			switch choice {
			case 0: // Обновить игру
				internal.ShowStyledMessage(internal.Info, "Начинается обновление игры...")
				err = internal.TryUnzipGame(gameDirPath, launcherPath)
				if err != nil {
					internal.ShowStyledMessage(internal.Error, "Ошибка: "+err.Error())
					continue
				}
				internal.ShowStyledMessage(internal.Success, "Игра успешно обновлена!")
				internal.TryRunGame(gameDirPath)
				shouldExit = true
			case 1: // Запустить игру
				internal.TryRunGame(gameDirPath)
				shouldExit = true
			case 2: // Выход
				shouldExit = true
			}
		} else {
			// Игра установлена и актуальна
			switch choice {
			case 0: // Запустить игру
				internal.TryRunGame(gameDirPath)
				shouldExit = true
			case 1: // Выход
				shouldExit = true
			}
		}

		if shouldExit {
			internal.ShowStyledMessage(internal.Info, "До свидания! 👋")
			return
		}
	}
}
