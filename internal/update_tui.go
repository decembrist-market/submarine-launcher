package internal

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UpdateModel - модель TUI для процесса обновления
type UpdateModel struct {
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

// NewUpdateModel создает новую модель обновления
func NewUpdateModel(gameDirPath, launcherPath string) UpdateModel {
	return UpdateModel{
		width:        80,
		height:       24,
		state:        StatePreparation,
		gameDirPath:  gameDirPath,
		launcherPath: launcherPath,
		progress:     InstallProgress{Current: 0, Total: 100, Message: "Подготовка к обновлению..."},
	}
}

func (m UpdateModel) Init() tea.Cmd {
	return tea.Batch(
		m.startUpdate(),
		m.tickCmd(),
	)
}

func (m UpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m UpdateModel) View() string {
	container := containerStyle.Width(m.width).Height(m.height)

	// Логотип
	logo := `🚢 ОБНОВЛЕНИЕ СУБМАРИНЫ 🚢`
	content := logoStyle.Width(m.width).Render(logo) + "\n\n"

	// Статус в зависимости от состояния
	switch m.state {
	case StatePreparation:
		statusMsg := fmt.Sprintf("%s Подготовка к обновлению...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateDownloading:
		statusMsg := fmt.Sprintf("%s Загрузка обновления...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateExtracting:
		statusMsg := fmt.Sprintf("%s Установка обновления...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateCompleted:
		content += installCompleteStyle.Width(m.width).Render("✅ Обновление завершено успешно!") + "\n\n"
		content += installStatusStyle.Width(m.width).Render("Игра будет запущена автоматически...") + "\n\n"

	case StateError:
		content += installErrorStyle.Width(m.width).Render("❌ Ошибка обновления") + "\n"
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

	// Центрирование для процесса обновления
	contentHeight := strings.Count(content, "\n") + 1
	emptyLines := (m.height - contentHeight) / 2
	if emptyLines < 0 {
		emptyLines = 0
	}

	result := strings.Repeat("\n", emptyLines) + content
	return container.Render(result)
}

func (m UpdateModel) renderProgressBar() string {
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

func (m UpdateModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m UpdateModel) startUpdate() tea.Cmd {
	return func() tea.Msg {
		// Имитация процесса обновления
		go func() {
			time.Sleep(time.Second * 1)
		}()
		return InstallProgressMsg{Current: 10, Total: 100, Message: "Очистка старых файлов..."}
	}
}

func (m UpdateModel) IsCompleted() bool {
	return m.completed
}

func (m UpdateModel) HasError() bool {
	return m.state == StateError
}

func (m UpdateModel) GetError() string {
	return m.errorMsg
}

// RunUpdateTUI запускает процесс обновления в TUI режиме
func RunUpdateTUI(gameDirPath, launcherPath string) error {
	model := NewUpdateModel(gameDirPath, launcherPath)

	// Создаем каналы для обновления прогресса
	progressChan := make(chan InstallProgress, 100)
	errorChan := make(chan error, 1)
	completeChan := make(chan bool, 1)

	// Запускаем программу TUI
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Запускаем обновление в горутине
	go func() {
		defer close(progressChan)
		defer close(errorChan)
		defer close(completeChan)

		// Обновление игры
		progressChan <- InstallProgress{Current: 10, Total: 100, Message: "Начало обновления..."}
		if err := installGameWithProgress(gameDirPath, launcherPath, progressChan); err != nil {
			errorChan <- err
			return
		}

		progressChan <- InstallProgress{Current: 100, Total: 100, Message: "Обновление завершено!"}
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

	updateModel := finalModel.(UpdateModel)
	if updateModel.HasError() {
		return fmt.Errorf("update failed: %s", updateModel.GetError())
	}

	return nil
}
