package internal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Стили для интерфейса
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D4AA")).
			Bold(true).
			Align(lipgloss.Center).
			Padding(1, 2)

	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CED1")).
			Bold(true).
			Align(lipgloss.Center)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 4)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00D4AA")).
				Bold(true).
				Padding(0, 4)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Align(lipgloss.Center).
			Padding(1, 0)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00D4AA")).
			Padding(1, 2).
			Align(lipgloss.Center)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#51CF66")).
			Bold(true)

	// Новые стили для полноэкранного режима
	containerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#0D1117")).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#484F58")).
			Align(lipgloss.Center).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("#21262D"))
)

type MenuChoice int

const (
	InstallGame MenuChoice = iota
	UpdateGame
	RunGame
	Exit
)

type TUIModel struct {
	choices       []string
	cursor        int
	gameInstalled bool
	needsUpdate   bool
	status        string
	statusType    string // "info", "error", "success"
	width         int    // Ширина терминала
	height        int    // Высота терминала
	selected      bool   // Был ли реально выбран пункт меню
}

func NewTUIModel(gameInstalled, needsUpdate bool) TUIModel {
	choices := []string{"🎮 Запустить игру", "🚪 Выход"}

	if !gameInstalled {
		choices = []string{"📦 Установить игру", "🚪 Выход"}
	} else if needsUpdate {
		choices = []string{"🔄 Обновить игру", "🎮 Запустить игру", "🚪 Выход"}
	}

	return TUIModel{
		choices:       choices,
		cursor:        0,
		gameInstalled: gameInstalled,
		needsUpdate:   needsUpdate,
		status:        "",
		statusType:    "info",
		width:         80, // Значение по умолчанию
		height:        24, // Значение по умолчанию
	}
}

func (m TUIModel) Init() tea.Cmd {
	return nil
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Обновляем размеры при изменении размера окна
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m TUIModel) View() string {
	// Создаем главный контейнер
	container := containerStyle.Width(m.width).Height(m.height)

	// ASCII лого
	logo := `🚢 СУБМАРИНА LAUNCHER 🚢`

	// Создаем основной контент
	content := ""

	// Добавляем логотип
	content += logoStyle.Width(m.width).Render(logo) + "\n\n"

	// Статус (показываем только если есть сообщение)
	if m.status != "" {
		var statusStyled string
		switch m.statusType {
		case "error":
			statusStyled = errorStyle.Render("❌ " + m.status)
		case "success":
			statusStyled = successStyle.Render("✅ " + m.status)
		default:
			statusStyled = statusStyle.Render("ℹ️  " + m.status)
		}
		content += lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(statusStyled) + "\n\n"
	}

	// Определяем состояние игры для отображения
	var gameStatus string
	if !m.gameInstalled {
		gameStatus = "🔴 Игра не установлена"
	} else if m.needsUpdate {
		gameStatus = "🟡 Доступно обновление"
	} else {
		gameStatus = "🟢 Игра готова к запуску"
	}

	// Отображаем статус игры
	statusBox := boxStyle.Width(m.width - 10).Render(gameStatus)
	content += lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(statusBox) + "\n\n"

	// Меню
	menuTitle := titleStyle.Width(m.width).Render("ВЫБЕРИТЕ ДЕЙСТВИЕ")
	content += menuTitle + "\n\n"

	// Рендерим меню по центру
	menu := ""
	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
			menu += selectedItemStyle.Width(30).Align(lipgloss.Center).Render(cursor+choice) + "\n"
		} else {
			menu += menuItemStyle.Width(30).Align(lipgloss.Center).Render(cursor+choice) + "\n"
		}
	}

	menuContainer := boxStyle.Width(40).Render(menu)
	content += lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(menuContainer)

	// Добавляем footer с подсказками
	footer := footerStyle.Width(m.width).Render("↑/↓ - навигация • Enter - выбрать • Esc/Q - выход")

	// Вычисляем сколько пустых строк нужно добавить для центрирования
	contentHeight := strings.Count(content, "\n") + 3 // +3 для footer
	emptyLines := (m.height - contentHeight) / 2
	if emptyLines < 0 {
		emptyLines = 0
	}

	// Собираем финальный результат
	result := strings.Repeat("\n", emptyLines) + content

	// Добавляем footer внизу
	footerPadding := m.height - strings.Count(result, "\n") - 2
	if footerPadding > 0 {
		result += strings.Repeat("\n", footerPadding)
	}
	result += footer

	return container.Render(result)
}

func (m TUIModel) GetChoice() int {
	return m.cursor
}

func (m TUIModel) WasSelected() bool {
	return m.selected
}

// Функция для отображения прогресса с красивым стилем
func ShowProgress(current, total float64, message string) {
	percent := int(current / total * 100)
	if percent > 100 {
		percent = 100
	}

	barWidth := 40
	filled := int(float64(barWidth) * float64(percent) / 100.0)

	progressBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00D4AA")).
		Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")).
			Render(strings.Repeat("░", barWidth-filled))

	progressText := fmt.Sprintf("%s [%s] %d%%", message, progressBar, percent)
	fmt.Printf("\r%s", progressText)

	if percent == 100 {
		fmt.Println()
	}
}

// Функция для отображения сообщения в красивом стиле
func ShowStyledMessage(msgType, message string) {
	var styledMsg string
	switch msgType {
	case "error":
		styledMsg = errorStyle.Render("❌ " + message)
	case "success":
		styledMsg = successStyle.Render("✅ " + message)
	case "warning":
		styledMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD43B")).
			Bold(true).
			Render("⚠️  " + message)
	default:
		styledMsg = statusStyle.Render("ℹ️  " + message)
	}

	fmt.Println(boxStyle.Render(styledMsg))
}

// Функция для подтверждения действия в красивом стиле
func ShowConfirmDialog(message string) bool {
	confirmBox := fmt.Sprintf("%s\n\n%s",
		message,
		statusStyle.Render("y/н для подтверждения, любая другая клавиша для отмены"))

	fmt.Println(boxStyle.Render(confirmBox))

	return CheckAnswer()
}
