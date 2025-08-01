package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// VersionInfo представляет новый формат версий
type VersionInfo struct {
	Version struct {
		Game     string `yaml:"game"`
		Launcher string `yaml:"launcher"`
	} `yaml:"version"`
}

// GetRemoteVersionInfo получает информацию о версиях с сервера
func GetRemoteVersionInfo() (*VersionInfo, error) {
	resp, err := http.Get(RemoteVersionURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе версии: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер вернул статус %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении ответа с сервера: %v", err)
	}

	var versionInfo VersionInfo
	err = yaml.Unmarshal(data, &versionInfo)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе YAML: %v", err)
	}

	return &versionInfo, nil
}

// GetLocalVersionInfo читает локальную версию игры
func GetLocalVersionInfo(versionFilePath string) (*VersionInfo, error) {
	data, err := os.ReadFile(versionFilePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла версии: %v", err)
	}

	var versionInfo VersionInfo
	err = yaml.Unmarshal(data, &versionInfo)
	if err != nil {
		// Попробуем старый формат для обратной совместимости
		oldVersion := strings.TrimSpace(string(data))
		if strings.HasPrefix(oldVersion, "version:") {
			oldVersion = strings.TrimPrefix(oldVersion, "version:")
			oldVersion = strings.TrimSpace(oldVersion)
		}

		versionInfo.Version.Game = oldVersion
		versionInfo.Version.Launcher = "0.0.0" // Старая версия не содержала версию лаунчера
	}

	return &versionInfo, nil
}

// NeedsLauncherUpdate проверяет, нужно ли обновить лаунчер
func NeedsLauncherUpdate(remoteVersion *VersionInfo) bool {
	return LauncherVersion != remoteVersion.Version.Launcher
}

// NeedsGameUpdate проверяет, нужно ли обновить игру
func NeedsGameUpdate(localVersion, remoteVersion *VersionInfo) bool {
	return localVersion.Version.Game != remoteVersion.Version.Game
}
