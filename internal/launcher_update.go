package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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
	tempLauncherPath := filepath.Join(dir, "SubmarineLauncher_new.exe")
	oldLauncherPath := filepath.Join(dir, "SubmarineLauncher_old.exe")

	// Скачиваем новую версию
	err := DownloadLauncherUpdate(tempLauncherPath)
	if err != nil {
		return err
	}

	// Создаем batch-скрипт для замены файлов
	batchPath := filepath.Join(dir, "update_launcher.bat")
	batchContent := fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
move "%s" "%s"
move "%s" "%s"
del "%s"
start "" "%s"
del "%%~f0"
`, currentLauncherPath, oldLauncherPath, tempLauncherPath, currentLauncherPath, oldLauncherPath, currentLauncherPath)

	err = os.WriteFile(batchPath, []byte(batchContent), 0755)
	if err != nil {
		os.Remove(tempLauncherPath) // Очищаем за собой
		return fmt.Errorf("ошибка при создании скрипта обновления: %v", err)
	}

	ShowStyledMessage(Info, "Запускаю обновление лаунчера...")

	// Запускаем batch-скрипт и завершаем текущий процесс
	cmd := exec.Command("cmd", "/c", batchPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008, // DETACHED_PROCESS
	}

	err = cmd.Start()
	if err != nil {
		os.Remove(tempLauncherPath)
		os.Remove(batchPath)
		return fmt.Errorf("ошибка при запуске скрипта обновления: %v", err)
	}

	ShowStyledMessage(Success, "Обновление запущено! Лаунчер будет перезапущен...")
	os.Exit(0) // Завершаем текущий процесс
	return nil
}
