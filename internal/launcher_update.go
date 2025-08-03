package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// DownloadLauncherUpdate скачивает новую версию лаунчера
func DownloadLauncherUpdate(tempPath string) error {
	ShowStyledMessage(Info, "Скачиваю обновление лаунчера...")

	resp, err := http.Get(GetLauncherURL())
	if err != nil {
		return fmt.Errorf("ошибка при скачивании лаунчера: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул статус %d при скачивании лаунчера", resp.StatusCode)
	}

	out, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании временного файла: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка при записи файла: %v", err)
	}

	ShowStyledMessage(Success, "Лаунчер успешно скачан!")
	return nil
}

// UpdateLauncher выполняет самообновление лаунчера
func UpdateLauncher(currentLauncherPath string) error {
	// Создаем временное имя для нового лаунчера
	dir := filepath.Dir(currentLauncherPath)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	tempLauncherPath := filepath.Join(dir, "SubmarineLauncher_new"+ext)
	oldLauncherPath := filepath.Join(dir, "SubmarineLauncher_old"+ext)

	// Скачиваем новую версию
	err := DownloadLauncherUpdate(tempLauncherPath)
	if err != nil {
		return err
	}

	// Создаем скрипт для замены файлов в зависимости от ОС
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
		os.Remove(tempLauncherPath) // Очищаем за собой
		return fmt.Errorf("ошибка при создании скрипта обновления: %v", err)
	}

	ShowStyledMessage(Info, "Запускаю обновление лаунчера...")

	// Запускаем скрипт в зависимости от ОС
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
		return fmt.Errorf("ошибка при запуске скрипта обновления: %v", err)
	}

	ShowStyledMessage(Success, "Обновление запущено! Лаунчер будет перезапущен...")
	os.Exit(0) // Завершаем текущий процесс
	return nil
}
