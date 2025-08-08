package internal

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LauncherUpdateModel - модель TUI для процесса обновления лаунчера
type LauncherUpdateModel struct {
	width        int
	height       int
	state        InstallState
	progress     InstallProgress
	errorMsg     string
	launcherPath string
	completed    bool
	spinner      int
	tickCount    int
}

// NewLauncherUpdateModel создает новую модель обновления лаунчера
func NewLauncherUpdateModel(launcherPath string) LauncherUpdateModel {
	return LauncherUpdateModel{
		width:        80,
		height:       24,
		state:        StatePreparation,
		launcherPath: launcherPath,
		progress:     InstallProgress{Current: 0, Total: 100, Message: "Подготовка к обновлению лаунчера..."},
	}
}

func (m LauncherUpdateModel) Init() tea.Cmd {
	return tea.Batch(
		m.startLauncherUpdate(),
		m.tickCmd(),
	)
}

func (m LauncherUpdateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.progress.Current >= 30 && m.state == StatePreparation {
			m.state = StateDownloading
		} else if m.progress.Current >= 80 && m.state == StateDownloading {
			m.state = StateExtracting
		} else if m.progress.Current >= 100 {
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

func (m LauncherUpdateModel) View() string {
	container := containerStyle.Width(m.width).Height(m.height)

	// Логотип
	logo := `🚀 ОБНОВЛЕНИЕ ЛАУНЧЕРА 🚀`
	content := logoStyle.Width(m.width).Render(logo) + "\n\n"

	// Статус в зависимости от состояния
	switch m.state {
	case StatePreparation:
		statusMsg := fmt.Sprintf("%s Подготовка к обновлению лаунчера...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateDownloading:
		statusMsg := fmt.Sprintf("%s Загрузка новой версии лаунчера...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateExtracting:
		statusMsg := fmt.Sprintf("%s Установка обновления лаунчера...", spinnerFrames[m.spinner])
		content += installStatusStyle.Width(m.width).Render(statusMsg) + "\n\n"

	case StateCompleted:
		content += installCompleteStyle.Width(m.width).Render("✅ Обновление лаунчера завершено!") + "\n\n"
		content += installStatusStyle.Width(m.width).Render("Лаунчер будет перезапущен автоматически...") + "\n\n"

	case StateError:
		content += installErrorStyle.Width(m.width).Render("❌ Ошибка обновления лаунчера") + "\n"
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

func (m LauncherUpdateModel) renderProgressBar() string {
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

func (m LauncherUpdateModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m LauncherUpdateModel) startLauncherUpdate() tea.Cmd {
	return func() tea.Msg {
		return InstallProgressMsg{Current: 0, Total: 100, Message: "Инициализация обновления..."}
	}
}

func (m LauncherUpdateModel) IsCompleted() bool {
	return m.completed
}

func (m LauncherUpdateModel) HasError() bool {
	return m.state == StateError
}

func (m LauncherUpdateModel) GetError() string {
	return m.errorMsg
}

// RunLauncherUpdateTUI запускает TUI для обновления лаунчера
func RunLauncherUpdateTUI(launcherPath string) error {
	model := NewLauncherUpdateModel(launcherPath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Запускаем горутину для реального обновления
	progressChan := make(chan InstallProgress, 10)
	errorChan := make(chan error, 1)
	completeChan := make(chan bool, 1)

	go func() {
		defer close(progressChan)
		defer close(errorChan)
		defer close(completeChan)

		err := updateLauncherWithProgress(launcherPath, progressChan)
		if err != nil {
			errorChan <- err
			return
		}

		completeChan <- true
	}()

	// Слушаем каналы и отправляем сообщения в TUI
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

	updateModel := finalModel.(LauncherUpdateModel)
	if updateModel.HasError() {
		return fmt.Errorf("launcher update failed: %s", updateModel.GetError())
	}

	return nil
}
