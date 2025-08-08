package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func runExecution(path string, logWriter io.Writer) error {
	cmd := exec.Command(path, "-launcher")

	// Создаем pipes для перехвата stdout и stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Запускаем команду
	if err := cmd.Start(); err != nil {
		return err
	}

	// Создаем multiwriter для записи в лог и stdout
	stdoutWriter := io.MultiWriter(logWriter, os.Stdout)
	stderrWriter := io.MultiWriter(logWriter, os.Stderr)

	// Копируем вывод в горутинах
	go io.Copy(stdoutWriter, stdout)
	go io.Copy(stderrWriter, stderr)

	// Ждем завершения команды
	return cmd.Wait()
}

// TryRunGame пытается запустить игру и ждет её завершения
func TryRunGame(dataDir string) error {
	fmt.Print("Запуск игры...\n")

	// Получаем имя исполняемого файла для текущей платформы
	gameFile := GetExecutableForPlatform()
	gamePath := filepath.Join(dataDir, gameFile)

	// Проверяем существование файла
	if _, err := os.Stat(gamePath); err != nil {
		ShowStyledMessage(Error, fmt.Sprintf("Исполняемый файл игры не найден: %s", gamePath))
		return fmt.Errorf("исполняемый файл игры не найден: %s", gamePath)
	}

	// На Unix-системах устанавливаем права на выполнение
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if err := os.Chmod(gamePath, 0755); err != nil {
			ShowStyledMessage(Error, "Ошибка при установке прав на выполнение: "+err.Error())
			return err
		}
	}

	// Создаем лог-файл с текущей датой и временем
	logDir := filepath.Join(filepath.Dir(dataDir), "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		ShowStyledMessage(Warn, "Не удалось создать папку для логов: "+err.Error())
		logDir = filepath.Dir(dataDir) // Используем родительскую папку
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := fmt.Sprintf("game_%s.log", timestamp)
	logPath := filepath.Join(logDir, logFileName)

	logFile, err := os.Create(logPath)
	if err != nil {
		ShowStyledMessage(Warn, "Не удалось создать лог-файл: "+err.Error())
		// Продолжаем без логирования
		logFile = nil
	} else {
		defer logFile.Close()
		// Записываем заголовок в лог
		fmt.Fprintf(logFile, "=== Лог игры начат: %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Fprintf(logFile, "Путь к игре: %s\n", gamePath)
		fmt.Fprintf(logFile, "============================\n\n")
	}

	var logWriter io.Writer = os.Stdout
	if logFile != nil {
		logWriter = logFile
		ShowStyledMessage(Info, fmt.Sprintf("Лог игры записывается в: %s", logPath))
	}

	ShowStyledMessage(Info, fmt.Sprintf("Игра запущена: %s", filepath.Base(gamePath)))

	err = runExecution(gamePath, logWriter)

	if logFile != nil {
		fmt.Fprintf(logFile, "\n============================\n")
		fmt.Fprintf(logFile, "=== Лог игры завершен: %s ===\n", time.Now().Format("2006-01-02 15:04:05"))
	}

	if err != nil {
		ShowStyledMessage(Error, "Игра завершилась с ошибкой: "+err.Error())
		return err
	}

	ShowStyledMessage(Info, "Игра завершена")
	return nil
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
