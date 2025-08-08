package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// updateLauncherWithProgress выполняет обновление лаунчера с отчетом о прогрессе
func updateLauncherWithProgress(currentLauncherPath string, progressChan chan<- InstallProgress) error {
	// Отправляем начальный прогресс
	progressChan <- InstallProgress{Current: 5, Total: 100, Message: "Подготовка к загрузке..."}

	// Определяем пути
	dir := filepath.Dir(currentLauncherPath)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	tempLauncherPath := filepath.Join(dir, "SubmarineLauncher_new"+ext)
	oldLauncherPath := filepath.Join(dir, "SubmarineLauncher_old"+ext)

	progressChan <- InstallProgress{Current: 10, Total: 100, Message: "Начинаем загрузку новой версии..."}

	// Загружаем новую версию лаунчера
	err := downloadLauncherUpdateWithProgress(tempLauncherPath, progressChan)
	if err != nil {
		return err
	}

	progressChan <- InstallProgress{Current: 80, Total: 100, Message: "Создание скрипта обновления..."}

	// Создаем и запускаем скрипт обновления
	var scriptPath string
	var scriptContent string

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(dir, "update_launcher.bat")
		scriptContent = fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
move "%s" "%s"
move "%s" "%s"
del "%s"
start "" "%s"
del "%%~f0"
`, currentLauncherPath, oldLauncherPath, tempLauncherPath, currentLauncherPath, oldLauncherPath, currentLauncherPath)
	} else {
		scriptPath = filepath.Join(dir, "update_launcher.sh")
		scriptContent = fmt.Sprintf(`#!/bin/bash
sleep 2
mv "%s" "%s"
mv "%s" "%s"
rm "%s"
"%s" &
rm "$0"
`, currentLauncherPath, oldLauncherPath, tempLauncherPath, currentLauncherPath, oldLauncherPath, currentLauncherPath)
	}

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		os.Remove(tempLauncherPath)
		return fmt.Errorf("ошибка при создании скрипта: %v", err)
	}

	progressChan <- InstallProgress{Current: 90, Total: 100, Message: "Запуск скрипта обновления..."}

	// Запускаем скрипт обновления
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", scriptPath)
	} else {
		cmd = exec.Command("sh", scriptPath)
	}

	err = cmd.Start()
	if err != nil {
		os.Remove(tempLauncherPath)
		os.Remove(scriptPath)
		return fmt.Errorf("ошибка при запуске скрипта: %v", err)
	}

	progressChan <- InstallProgress{Current: 100, Total: 100, Message: "Обновление завершено!"}

	// Небольшая задержка перед завершением процесса
	time.Sleep(2 * time.Second)
	os.Exit(0) // Завершаем текущий процесс
	return nil
}

// downloadLauncherUpdateWithProgress загружает обновление лаунчера с прогрессом
func downloadLauncherUpdateWithProgress(tempPath string, progressChan chan<- InstallProgress) error {
	progressChan <- InstallProgress{Current: 15, Total: 100, Message: "Подключение к серверу..."}

	resp, err := http.Get(GetLauncherURL())
	if err != nil {
		return fmt.Errorf("ошибка при загрузке обновления: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный статус ответа %d при загрузке обновления", resp.StatusCode)
	}

	progressChan <- InstallProgress{Current: 20, Total: 100, Message: "Начинаем загрузку файла..."}

	out, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %v", err)
	}
	defer out.Close()

	// Создаем прогресс-ридер для отслеживания загрузки
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		contentLength = 10 * 1024 * 1024 // Предполагаем 10MB если размер неизвестен
	}

	var written int64
	buffer := make([]byte, 32*1024) // 32KB буфер

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := out.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("ошибка при записи файла: %v", writeErr)
			}
			written += int64(n)

			// Обновляем прогресс (загрузка занимает 20-75% от общего прогресса)
			percent := float64(written) / float64(contentLength)
			if percent > 1.0 {
				percent = 1.0
			}
			current := 20 + int(55*percent)
			progressChan <- InstallProgress{
				Current: current,
				Total:   100,
				Message: fmt.Sprintf("Загружено: %.1f MB", float64(written)/1024/1024),
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка при чтении файла: %v", err)
		}
	}

	progressChan <- InstallProgress{Current: 75, Total: 100, Message: "Загрузка завершена!"}
	return nil
}

// DownloadLauncherUpdate загружает обновление лаунчера (старая функция для совместимости)
func DownloadLauncherUpdate(tempPath string) error {
	ShowStyledMessage(Info, "Загрузка обновления...")

	resp, err := http.Get(GetLauncherURL())
	if err != nil {
		return fmt.Errorf("ошибка при загрузке обновления: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("неожиданный статус ответа %d при загрузке обновления", resp.StatusCode)
	}

	out, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при записи файла: %v", err)
	}

	ShowStyledMessage(Success, "Обновление скачано успешно!")
	return nil
}

// UpdateLauncher выполняет самообновление лаунчера (старая функция для совместимости)
func UpdateLauncher(currentLauncherPath string) error {
	// Определяем пути для файлов
	dir := filepath.Dir(currentLauncherPath)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	tempLauncherPath := filepath.Join(dir, "SubmarineLauncher_new"+ext)
	oldLauncherPath := filepath.Join(dir, "SubmarineLauncher_old"+ext)

	// Загружаем обновление
	err := DownloadLauncherUpdate(tempLauncherPath)
	if err != nil {
		return err
	}

	// Создаем и записываем скрипт для замены файлов
	var scriptPath string
	var scriptContent string

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(dir, "update_launcher.bat")
		scriptContent = fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
move "%s" "%s"
move "%s" "%s"
del "%s"
start "" "%s"
del "%%~f0"
`, currentLauncherPath, oldLauncherPath, tempLauncherPath, currentLauncherPath, oldLauncherPath, currentLauncherPath)
	} else {
		scriptPath = filepath.Join(dir, "update_launcher.sh")
		scriptContent = fmt.Sprintf(`#!/bin/bash
sleep 2
mv "%s" "%s"
mv "%s" "%s"
rm "%s"
"%s" &
rm "$0"
`, currentLauncherPath, oldLauncherPath, tempLauncherPath, currentLauncherPath, oldLauncherPath, currentLauncherPath)
	}

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		os.Remove(tempLauncherPath) // Очищаем файл
		return fmt.Errorf("ошибка при создании скрипта: %v", err)
	}

	ShowStyledMessage(Info, "Выполняем обновление...")

	// Запускаем скрипт и завершаем процесс
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", scriptPath)
	} else {
		cmd = exec.Command("sh", scriptPath)
	}

	err = cmd.Start()
	if err != nil {
		os.Remove(tempLauncherPath)
		os.Remove(scriptPath)
		return fmt.Errorf("ошибка запуска скрипта: %v", err)
	}

	ShowStyledMessage(Success, "Обновление запущено! Лаунчер перезапустится...")
	os.Exit(0) // Завершаем текущий процесс
	return nil
}
