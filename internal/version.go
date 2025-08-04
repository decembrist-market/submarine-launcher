package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// parseVersion разбирает версию в формате "major.minor.patch[-suffix]"
func parseVersion(version string) ([]int, string, error) {
	// Разделяем основную версию и суффикс
	var versionPart, suffix string
	if dashIndex := strings.Index(version, "-"); dashIndex != -1 {
		versionPart = version[:dashIndex]
		suffix = version[dashIndex+1:]
	} else {
		versionPart = version
		suffix = ""
	}

	parts := strings.Split(versionPart, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return nil, "", fmt.Errorf("неверный формат версии: %s", version)
	}

	var nums []int
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, "", fmt.Errorf("неверный формат версии: %s", version)
		}
		nums = append(nums, num)
	}

	// Дополняем до 3 компонентов нулями если нужно
	for len(nums) < 3 {
		nums = append(nums, 0)
	}

	return nums, suffix, nil
}

// getSuffixPriority возвращает приоритет суффикса для сравнения
// Чем больше число, тем выше приоритет
func getSuffixPriority(suffix string) int {
	switch suffix {
	case "alpha":
		return 1
	case "beta":
		return 2
	case "": // релиз (без суффикса)
		return 3
	default:
		return 0 // неизвестный суффикс имеет самый низкий приоритет
	}
}

// CompareVersions сравнивает две версии семантически с учетом суффиксов
// Возвращает: -1 если v1 < v2, 0 если v1 == v2, 1 если v1 > v2
func CompareVersions(v1, v2 string) (int, error) {
	nums1, suffix1, err := parseVersion(v1)
	if err != nil {
		return 0, err
	}

	nums2, suffix2, err := parseVersion(v2)
	if err != nil {
		return 0, err
	}

	// Сначала сравниваем основные компоненты версии
	for i := 0; i < 3; i++ {
		if nums1[i] < nums2[i] {
			return -1, nil
		} else if nums1[i] > nums2[i] {
			return 1, nil
		}
	}

	// Если основные версии равны, сравниваем суффиксы
	priority1 := getSuffixPriority(suffix1)
	priority2 := getSuffixPriority(suffix2)

	if priority1 < priority2 {
		return -1, nil
	} else if priority1 > priority2 {
		return 1, nil
	}

	return 0, nil
}

// IsVersionNewer проверяет, является ли remoteVersion новее localVersion
func IsVersionNewer(localVersion, remoteVersion string) (bool, error) {
	result, err := CompareVersions(localVersion, remoteVersion)
	if err != nil {
		return false, err
	}
	return result < 0, nil
}

// CustomTime представляет время, которое можно разобрать из более короткого формата
type CustomTime struct {
	time.Time
}

// UnmarshalYAML реализует пользовательскую разборку для форматов времени
func (ct *CustomTime) UnmarshalYAML(value *yaml.Node) error {
	var timeStr string
	if err := value.Decode(&timeStr); err != nil {
		return err
	}

	// Пробуем разобрать с разными форматами
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // Полный RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 без смещения часового пояса
		"2006-01-02T15:04:05",       // Без часового пояса
		"2006-01-02T15:04",          // Без секунд и часового пояса
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			ct.Time = t
			return nil
		}
	}

	return fmt.Errorf("не удалось разобрать время: %s", timeStr)
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
	isNewer, err := IsVersionNewer(LauncherVersion, remoteVersion.Version.Launcher)
	if err != nil {
		// Если не удалось сравнить версии семантически, используем строковое сравнение
		return LauncherVersion != remoteVersion.Version.Launcher
	}
	return isNewer
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
