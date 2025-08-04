# Submarine Launcher

Кроссплатформенный лаунчер для игры Submarine Game с автоматическим обновлением и красивым терминальным интерфейсом.

## Возможности

- 🚀 Автоматическое обновление игры и самого лаунчера
- 🎨 Красивый терминальный интерфейс с прогресс-барами
- 🔍 Проверка целостности файлов через SHA256
- 📱 Поддержка Windows, Linux и macOS
- 🛠️ Обработка технического обслуживания
- 📢 Серверные сообщения для пользователей
- 🌈 Поддержка цветов в Windows Terminal

## Системные требования

- Go 1.24 или выше
- Интернет-соединение для загрузки обновлений
- Поддерживаемые ОС: Windows, Linux, macOS

## Сборка проекта

### Клонирование репозитория

```bash
git clone <repository-url>
cd submarine-launcher
```

### Установка зависимостей

```bash
go mod download
```

### Сборка для текущей платформы

```bash
go build -o SubmarineLauncher main.go
```

### Сборка для Windows

```bash
# Из Windows
go build -o SubmarineLauncher.exe main.go

# Кроссплатформенная сборка
GOOS=windows GOARCH=amd64 go build -o SubmarineLauncher.exe main.go
```

### Сборка для Linux

```bash
# Из Linux
go build -o SubmarineLauncher main.go

# Кроссплатформенная сборка
GOOS=linux GOARCH=amd64 go build -o SubmarineLauncher main.go
```

### Сборка для macOS

```bash
# Из macOS
go build -o SubmarineLauncher main.go

# Кроссплатформенная сборка
GOOS=darwin GOARCH=amd64 go build -o SubmarineLauncher main.go
```

### Сборка для всех платформ

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o export/SubmarineLauncher.exe main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o export/SubmarineLauncher main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o export/SubmarineLauncher main.go
```

## Запуск

После сборки просто запустите исполняемый файл:

```bash
# Windows
SubmarineLauncher.exe

# Linux/macOS
./SubmarineLauncher
```

### Используемые библиотеки

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI фреймворк
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Стилизация терминального интерфейса
- **[yaml.v3](https://gopkg.in/yaml.v3)** - Парсинг YAML файлов

### Запуск в режиме разработки

```bash
go run main.go
```

### Тестирование

```bash
go test ./...
```

## Конфигурация

Лаунчер использует удаленный манифест для проверки версий и получения настроек:

- **URL манифеста**: `https://static.decembrist.org/submarine-game/launcher-manifest.yaml`
- **Версия лаунчера**: `0.0.2`
- **Папка игры**: `SubmarineGame`

## Функциональность

### Автоматическое обновление

Лаунчер автоматически проверяет обновления при запуске:

1. Проверяет версию лаунчера и при необходимости обновляется
2. Проверяет версию игры и предлагает обновление
3. Проверяет целостность файлов игры

### Техническое обслуживание

Поддерживается планирование технического обслуживания через манифест:

- Предупреждения о предстоящем обслуживании
- Блокировка запуска игры во время обслуживания

### Серверные сообщения

Возможность отображения сообщений пользователям:

- Обычные предупреждения (желтый цвет)
- Критические сообщения (красный цвет)
