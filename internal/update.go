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

func TryUnzipGame(dir, updaterPath string) error {
	err := removeOldFiles(dir, updaterPath)
	if err != nil {
		return fmt.Errorf("ошибка при удалении старых файлов: %v", err)
	}

	archiveFile, err := os.CreateTemp("", ArchiveNameTemplate)
	if err != nil {
		return fmt.Errorf("ошибка при создании временного файла архива: %v", err)
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
		return fmt.Errorf("ошибка при загрузке архива: %v", err)
	}

	//todo
	//err = checkHash(archivePath)
	//if err != nil {
	//	return fmt.Errorf("ошибка при проверке хеша архива: %v", err)
	//}
	//ShowStyledMessage(Info, "Хеш архива успешно проверен")

	err = unzipWithProgress(archivePath, dir)
	if err != nil {
		return fmt.Errorf("ошибка при распаковке архива: %v", err)
	}
	return nil
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
	ShowStyledMessage(Info, "Загрузка архива игры...")
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
	ShowStyledMessage(Success, "Загрузка завершена!")
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

	ShowStyledMessage(Info, "Распаковка архива...")
	for i, file := range reader.File {
		filePath := filepath.Join(dir, file.Name)

		ShowProgress(float64(i), float64(totalFiles), "📦 Распаковываем")

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

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		targetFile, err := os.Create(filePath)
		if err != nil {
			fileReader.Close()
			return err
		}

		_, err = io.Copy(targetFile, fileReader)
		targetFile.Close()
		fileReader.Close()

		if err != nil {
			return err
		}
	}

	fmt.Println()
	ShowStyledMessage(Success, "Распаковка завершена!")
	return nil
}

// downloadZipWithProgress загружает архив с отправкой прогресса в TUI
func downloadZipWithProgress(archiveFile *os.File, progressChan chan<- InstallProgress) error {
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
	downloaded := int64(0)

	for {
		readBytes, err := resp.Body.Read(buf)
		if readBytes > 0 {
			_, err2 := archiveFile.Write(buf[:readBytes])
			if err2 != nil {
				return err2
			}
			downloaded += int64(readBytes)

			// Рассчитываем прогресс (25-70%)
			percent := int(float64(downloaded)/float64(total)*45) + 25
			if percent > 70 {
				percent = 70
			}

			progressChan <- InstallProgress{
				Current: percent,
				Total:   100,
				Message: fmt.Sprintf("Загружено: %.1f MB / %.1f MB",
					float64(downloaded)/(1024*1024),
					float64(total)/(1024*1024)),
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("ошибка при чтении данных: %v", err)
		}
	}

	return nil
}

// unzipWithProgressTUI распаковывает архив с отправкой прогресса в TUI
func unzipWithProgressTUI(src, dir string, progressChan chan<- InstallProgress) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	totalFiles := len(reader.File)
	if totalFiles == 0 {
		totalFiles = 1
	}

	for i, file := range reader.File {
		filePath := filepath.Join(dir, file.Name)

		// Рассчитываем прогресс (70-95%)
		percent := int(float64(i)/float64(totalFiles)*25) + 70
		if percent > 95 {
			percent = 95
		}

		progressChan <- InstallProgress{
			Current: percent,
			Total:   100,
			Message: fmt.Sprintf("Распаковка: %d/%d файлов", i+1, totalFiles),
		}

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

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		targetFile, err := os.Create(filePath)
		if err != nil {
			fileReader.Close()
			return err
		}

		_, err = io.Copy(targetFile, fileReader)
		targetFile.Close()
		fileReader.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
