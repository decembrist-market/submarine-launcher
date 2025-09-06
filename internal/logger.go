package internal

import (
	"fmt"
	"io"
)

type Logger struct {
	writer   io.Writer
	lastLine string
	dupCount int
}

func CreateLogger(w io.Writer) *Logger {
	return &Logger{writer: w}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	line := string(p)
	if line == l.lastLine {
		l.dupCount++
		return len(p), nil
	}

	if l.dupCount > 0 {
		fmt.Fprintf(l.writer, "[REPEAT] The message↑ was repeated %d times\n", l.dupCount+1)
		l.dupCount = 0
	}
	l.lastLine = line
	fmt.Fprint(l.writer, line)

	return len(p), nil
}

func (l *Logger) FlushRepeat() {
	if l.lastLine != "" {
		if l.dupCount > 0 {
			fmt.Fprintf(l.writer, "[REPEAT] %s - was repeated %d times\n", l.lastLine, l.dupCount+1)
		}
	}
}
