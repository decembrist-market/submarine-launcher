package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const ArchiveNameTemplate = "submarine-archive-*.zip"

func GetRemoteVersion() (string, error) {
	resp, err := http.Get(RemoteVersionURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при запросе версии: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("сервер вернул статус %d", resp.StatusCode)
	}
	remoteVersion, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа с сервера: %v", err)
	}
	return string(remoteVersion), nil
}

func removeOldFiles(dir, launcherPath string) error {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ошибка при чтении директории %s: %v", dir, err)
	}
	for _, entry := range dirEntries {
		entryPath := filepath.Join(dir, entry.Name())
		if entryPath == launcherPath {
			continue
		}
		err := os.RemoveAll(entryPath)
		if err != nil {
			return fmt.Errorf("ошибка при удалении файла %s: %v", entryPath, err)
		} else {
			fmt.Printf("Удалено: %s\n", entryPath)
		}
	}
	return nil
}

func TryUnzipGame(dir, updaterPath string) {
	err := removeOldFiles(dir, updaterPath)
	if err != nil {
		fmt.Println("Ошибка при удалении файлов игры:", err)
		return
	}

	archiveFile, err := os.CreateTemp("", ArchiveNameTemplate)
	if err != nil {
		fmt.Println("Ошибка при создании временного файла архива:", err)
		return
	}

	archivePath := archiveFile.Name()
	defer func() {
		archiveFile.Close()
		err = os.Remove(archivePath)
		if err != nil {
			fmt.Println("Ошибка при удалении временного файла архива:", err)
		}
	}()

	err = downloadZip(archiveFile)
	if err != nil {
		fmt.Println("Ошибка при загрузке архива:", err)
		return
	}

	err = unzipWithProgress(archivePath, dir)
	if err != nil {
		fmt.Println("Ошибка при распаковке архива:", err)
		return
	}
}

func downloadZip(archiveFile *os.File) error {
	resp, err := http.Get(ArchiveURL)
	if err != nil {
		return fmt.Errorf("ошибка при загрузке архива: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("сервер вернул статус %d", resp.StatusCode)
	}

	total := resp.ContentLength
	if total <= 0 {
		total = 1
	}

	buf := make([]byte, 32*1024)
	downloaded := 0.0
	ShowStyledMessage("info", "Загрузка архива игры...")
	for {
		readBytes, err := resp.Body.Read(buf)
		if readBytes > 0 {
			_, err2 := archiveFile.Write(buf[:readBytes])
			if err2 != nil {
				return err2
			}
			downloaded += float64(readBytes)
			ShowProgress(downloaded, float64(total), "📦 Загружаем")
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка при чтении данных: %v", err)
		}
	}
	fmt.Println()
	ShowStyledMessage("success", "Загрузка завершена!")
	return nil
}

func unzipWithProgress(src, dir string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()
	totalFiles := len(reader.File)
	if totalFiles == 0 {
		totalFiles = 1
	}
	ShowStyledMessage("info", "Распаковка архива...")
	for i, file := range reader.File {
		filePath := filepath.Join(dir, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		readCloser, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, readCloser)
		readCloser.Close()
		outFile.Close()
		if err != nil {
			return err
		}
		ShowProgress(float64(i+1), float64(totalFiles), "📂 Распаковываем")
	}
	fmt.Println()
	ShowStyledMessage("success", "Распаковка завершена!")
	return nil
}
