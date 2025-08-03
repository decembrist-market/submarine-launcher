package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstallState представляет состояния процесса установки
type InstallState int

const (
	StatePreparation InstallState = iota
	StateDownloading
	StateExtracting
	StateCompleted
	StateError
)

// InstallProgress представляет прогресс установки
type InstallProgress struct {
	Current int
	Total   int
	Message string
}

// InstallModel - модель TUI для процесса установки
type InstallModel struct {
	width        int
	height       int
	state        InstallState
	progress     InstallProgress
	errorMsg     string
	gameDirPath  string
	launcherPath string
	completed    bool
	spinner      int
	tickCount    int
}

// Стили для установки
var (
	installTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D4AA")).
				Bold(true).
				Align(lipgloss.Center).
				Padding(1, 2)

	installProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D4AA")).
				Bold(true)

	installProgressBgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#333333"))

	installStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Align(lipgloss.Center).
				Padding(1, 0)

	installErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B")).
				Bold(true).
				Align(lipgloss.Center)

	installCompleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#51CF66")).
				Bold(true).
				Align(lipgloss.Center)
)

// Символы спиннера
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewInstallModel создает новую модель установки
func NewInstallModel(gameDirPath, launcherPath string) InstallModel {
	return InstallModel{
		width:        80,
		height:       24,
		state:        StatePreparation,
		gameDirPath:  gameDirPath,
		launcherPath: launcherPath,
		progress:     InstallProgress{Current: 0, Total: 100, Message: "Подготовка..."},
	}
}

// Сообщения для обновления модели
type InstallProgressMsg InstallProgress
type InstallErrorMsg string
type InstallCompleteMsg struct{}
type TickMsg time.Time

func (m InstallModel) Init() tea.Cmd {
	return tea.Batch(
		m.startInstallation(),
		m.tickCmd(),
	)
}

func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		m.tickCount++
		m.spinner = (m.spinner + 1) % len(spinnerFrames)
		if m.state != StateCompleted && m.state != StateError {
			return m, m.tickCmd()
		}
		return m, nil

	case InstallProgressMsg:
		m.progress = InstallProgress(msg)
		if m.progress.Current >= 50 && m.state == StateDownloading {
			m.state = StateExtracting
		} else if m.progress.Current >= 90 && m.state == StateExtracting {
			m.state = StateCompleted
		}
		return m, nil

	case InstallErrorMsg:
		m.state = StateError
		m.errorMsg = string(msg)
		return m, nil

	case InstallCompleteMsg:
		m.state = StateCompleted
		m.completed = true
		return m, nil

	case tea.KeyMsg:
		if m.state == StateCompleted || m.state == StateError {
			switch msg.String() {
			case "enter", " ":
				return m, tea.Quit
			case "ctrl+c", "q", "esc":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m InstallModel) View() string {
	container := containerStyle.Width(m.width).Height(m.height)

	// Логотип
	logo := `🚢 УСТАНОВКА СУБМАРИНЫ 🚢`
	content := logoStyle.Width(m.width).Render(logo) + "\n\n"

	// Статус в зависимости от состояния
	switch m.state {
	case StatePreparation:
		statusMsg := fmt.Sprintf("%s Подготовка к установке...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateDownloading:
		statusMsg := fmt.Sprintf("%s Загрузка игры...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateExtracting:
		statusMsg := fmt.Sprintf("%s Распаковка файлов...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateCompleted:
		content += installCompleteStyle.Width(m.width).Render("✅ Установка завершена успешно!") + "\n\n"

	case StateError:
		content += installErrorStyle.Width(m.width).Render("❌ Ошибка установки") + "\n"
		content += installErrorStyle.Width(m.width).Render(m.errorMsg) + "\n\n"
	}

	// Прогресс бар
	if m.state != StateError {
		progressBar := m.renderProgressBar()
		content += lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(progressBar) + "\n\n"

		// Сообщение о прогрессе
		content += installStatusStyle.Width(m.width).Render(m.progress.Message) + "\n\n"
	}

	// Инструкции
	if m.state == StateCompleted || m.state == StateError {
		footer := footerStyle.Width(m.width).Render("Нажмите Enter для продолжения")
		contentHeight := strings.Count(content, "\n") + 3
		emptyLines := (m.height - contentHeight) / 2
		if emptyLines < 0 {
			emptyLines = 0
		}

		result := strings.Repeat("\n", emptyLines) + content
		footerPadding := m.height - strings.Count(result, "\n") - 2
		if footerPadding > 0 {
			result += strings.Repeat("\n", footerPadding)
		}
		result += footer
		return container.Render(result)
	}

	// Центрирование для процесса установки
	contentHeight := strings.Count(content, "\n") + 1
	emptyLines := (m.height - contentHeight) / 2
	if emptyLines < 0 {
		emptyLines = 0
	}

	result := strings.Repeat("\n", emptyLines) + content
	return container.Render(result)
}

func (m InstallModel) renderProgressBar() string {
	barWidth := 50
	percent := float64(m.progress.Current) / float64(m.progress.Total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(float64(barWidth) * percent)
	progressBar := installProgressStyle.Render(strings.Repeat("█", filled)) +
		installProgressBgStyle.Render(strings.Repeat("░", barWidth-filled))

	percentText := fmt.Sprintf(" %d%%", int(percent*100))
	return fmt.Sprintf("[%s]%s", progressBar, percentText)
}

func (m InstallModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m InstallModel) startInstallation() tea.Cmd {
	return func() tea.Msg {
		// Имитация процесса установки
		go func() {
			// Здесь будет вызов реальной функции установки
			// Пока что имитация
			time.Sleep(time.Second * 1)
		}()
		return InstallProgressMsg{Current: 10, Total: 100, Message: "Создание директории..."}
	}
}

func (m InstallModel) IsCompleted() bool {
	return m.completed
}

func (m InstallModel) HasError() bool {
	return m.state == StateError
}

func (m InstallModel) GetError() string {
	return m.errorMsg
}

// RunInstallationTUI запускает процесс установки в TUI режиме
func RunInstallationTUI(gameDirPath, launcherPath string) error {
	model := NewInstallModel(gameDirPath, launcherPath)

	// Создаем канал для обновления прогресса
	progressChan := make(chan InstallProgress, 100)
	errorChan := make(chan error, 1)
	completeChan := make(chan bool, 1)

	// Запускаем программу TUI
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Запускаем установку в горутине
	go func() {
		defer close(progressChan)
		defer close(errorChan)
		defer close(completeChan)

		// Создание директории
		progressChan <- InstallProgress{Current: 5, Total: 100, Message: "Создание директории игры..."}
		if err := createGameDirectory(gameDirPath); err != nil {
			errorChan <- err
			return
		}

		// Загрузка и установка
		progressChan <- InstallProgress{Current: 10, Total: 100, Message: "Начало загрузки..."}
		if err := installGameWithProgress(gameDirPath, launcherPath, progressChan); err != nil {
			errorChan <- err
			return
		}

		progressChan <- InstallProgress{Current: 100, Total: 100, Message: "Установка завершена!"}
		completeChan <- true
	}()

	// Запускаем TUI с обработкой сообщений
	go func() {
		for {
			select {
			case progress, ok := <-progressChan:
				if !ok {
					progressChan = nil
					continue
				}
				p.Send(InstallProgressMsg(progress))
			case err, ok := <-errorChan:
				if !ok {
					errorChan = nil
					continue
				}
				if err != nil {
					p.Send(InstallErrorMsg(err.Error()))
				}
			case _, ok := <-completeChan:
				if !ok {
					completeChan = nil
					continue
				}
				p.Send(InstallCompleteMsg{})
			}

			if progressChan == nil && errorChan == nil && completeChan == nil {
				break
			}
		}
	}()

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	installModel := finalModel.(InstallModel)
	if installModel.HasError() {
		return fmt.Errorf("installation failed: %s", installModel.GetError())
	}

	return nil
}

// createGameDirectory создает директорию для игры
func createGameDirectory(gameDirPath string) error {
	if _, err := os.Stat(gameDirPath); os.IsNotExist(err) {
		return os.Mkdir(gameDirPath, 0755)
	}
	return nil
}

// installGameWithProgress выполняет установку игры с отправкой прогресса
func installGameWithProgress(gameDirPath, launcherPath string, progressChan chan<- InstallProgress) error {
	// Удаление старых файлов
	progressChan <- InstallProgress{Current: 15, Total: 100, Message: "Очистка старых файлов..."}
	if err := removeOldFilesQuiet(gameDirPath, launcherPath); err != nil {
		return fmt.Errorf("ошибка при удалении старых файлов: %v", err)
	}

	// Создание временного файла
	progressChan <- InstallProgress{Current: 20, Total: 100, Message: "Подготовка к загрузке..."}
	archiveFile, err := os.CreateTemp("", ArchiveNameTemplate)
	if err != nil {
		return fmt.Errorf("ошибка при создании временного файла архива: %v", err)
	}

	archivePath := archiveFile.Name()
	defer func() {
		archiveFile.Close()
		os.Remove(archivePath)
	}()

	// Загрузка архива
	progressChan <- InstallProgress{Current: 25, Total: 100, Message: "Загрузка архива игры..."}
	if err := downloadZipWithProgress(archiveFile, progressChan); err != nil {
		return fmt.Errorf("ошибка при загрузке архива: %v", err)
	}

	// Распаковка архива
	progressChan <- InstallProgress{Current: 70, Total: 100, Message: "Распаковка файлов игры..."}
	if err := unzipWithProgressTUI(archivePath, gameDirPath, progressChan); err != nil {
		return fmt.Errorf("ошибка при распаковке архива: %v", err)
	}

	return nil
}

// removeOldFilesQuiet - тихая версия removeOldFiles без консольного вывода
func removeOldFilesQuiet(dir, launcherPath string) error {
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
		}
		// Убираем консольный вывод для TUI режима
	}
	return nil
}
