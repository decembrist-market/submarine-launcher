package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// CustomTime represents a time that can be unmarshaled from shorter format
type CustomTime struct {
	time.Time
}

// UnmarshalYAML implements custom unmarshaling for time formats
func (ct *CustomTime) UnmarshalYAML(value *yaml.Node) error {
	var timeStr string
	if err := value.Decode(&timeStr); err != nil {
		return err
	}

	// Try parsing with different formats
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // Full RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 without timezone offset
		"2006-01-02T15:04:05",       // Without timezone
		"2006-01-02T15:04",          // Without seconds and timezone
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			ct.Time = t
			return nil
		}
	}

	return fmt.Errorf("unable to parse time: %s", timeStr)
}

// ManifestDto представляет новый формат версий
type ManifestDto struct {
	Version struct {
		Game     string `yaml:"game"`
		Launcher string `yaml:"launcher"`
	} `yaml:"version"`
	Shutdown *CustomTime `yaml:"shutdown,omitempty"`
	Message  *struct {
		Text      string `yaml:"text"`
		Important bool   `yaml:"important"`
	} `yaml:"message,omitempty"`
}

// GetRemoteManifest получает информацию о версиях с сервера
func GetRemoteManifest() (*ManifestDto, error) {
	resp, err := http.Get(RemoteManifestURL)
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

	var manifest ManifestDto
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе YAML: %v", err)
	}

	return &manifest, nil
}

// GetGameLocalVersion читает локальную версию игры
func GetGameLocalVersion(versionFilePath string) (string, error) {
	data, err := os.ReadFile(versionFilePath)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении файла версии: %v", err)
	}

	// GameVersionDto представляет формат версии игры
	type GameVersionDto struct {
		Version string `yaml:"version"`
	}

	var gameVersion *GameVersionDto
	err = yaml.Unmarshal(data, &gameVersion)
	if err != nil {
		return "", fmt.Errorf("ошибка при разборе YAML версии игры: %v", err)
	}

	return gameVersion.Version, nil
}

// NeedsLauncherUpdate проверяет, нужно ли обновить лаунчер
func NeedsLauncherUpdate(remoteVersion *ManifestDto) bool {
	return LauncherVersion != remoteVersion.Version.Launcher
}

// NeedsGameUpdate проверяет, нужно ли обновить игру
func NeedsGameUpdate(localVersion, remoteVersion *ManifestDto) bool {
	return localVersion.Version.Game != remoteVersion.Version.Game
}

// GetMaintenanceMessage возвращает сообщение о техническом обслуживании
func GetMaintenanceMessage(versionInfo *ManifestDto) (string, string) {
	if versionInfo.Shutdown == nil {
		return "", ""
	}

	now := time.Now().UTC()
	shutdownTime := versionInfo.Shutdown.Time

	if now.Before(shutdownTime) {
		// Техническое обслуживание еще не началось
		return fmt.Sprintf("Внимание! В %s начнется техническое обслуживание, в это время игра будет недоступна",
			shutdownTime.Format("2006-01-02 15:04")), Warn
	} else {
		// Техническое обслуживание уже идет
		return "Внимание! Идет техническое обслуживание, игра недоступна", Error
	}
}

// GetServerMessage возвращает серверное сообщение
func GetServerMessage(versionInfo *ManifestDto) (string, string) {
	if versionInfo.Message == nil || versionInfo.Message.Text == "" {
		return "", ""
	}

	messageType := Warn
	if versionInfo.Message.Important {
		messageType = Error
	}

	return versionInfo.Message.Text, messageType
}

// IsGameAccessible проверяет, доступна ли игра (не идет ли техническое обслуживание)
func IsGameAccessible(versionInfo *ManifestDto) bool {
	if versionInfo.Shutdown == nil {
		return true
	}

	now := time.Now().UTC()
	return now.Before(versionInfo.Shutdown.Time)
}
