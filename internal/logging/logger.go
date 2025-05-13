package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// Logger 封装日志输出：支持 Info / Debug / Success / Warning / Error
type Logger struct {
	Verbose bool
	buf     *bytes.Buffer
	logger  *log.Logger
}

// ANSI 控制台颜色定义
var (
	gray   = "\033[90m"
	blue   = "\033[94m"
	green  = "\033[92m"
	yellow = "\033[93m"
	red    = "\033[91m"
	reset  = "\033[0m"
)

// 创建新的 Logger 实例
func NewLogger(verbose bool) *Logger {
	buf := new(bytes.Buffer)
	multi := io.MultiWriter(os.Stdout, buf)
	logger := log.New(multi, "", log.LstdFlags)
	return &Logger{Verbose: verbose, buf: buf, logger: logger}
}

// 普通信息
func (l *Logger) Info(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Println(blue + "ℹ️ " + msg + reset)
}

// 调试信息（仅 verbose=true 时输出）
func (l *Logger) Debug(format string, v ...any) {
	if l.Verbose {
		msg := fmt.Sprintf(format, v...)
		l.logger.Println(gray + "🔍 " + msg + reset)
	}
}

// 成功信息（绿色）
func (l *Logger) Success(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Println(green + "✅ " + msg + reset)
}

// 警告信息（黄色）
func (l *Logger) Warning(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Println(yellow + "⚠️  " + msg + reset)
}

// 错误信息（红色）
func (l *Logger) Error(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	l.logger.Println(red + "❌ " + msg + reset)
}

// 获取日志输出缓存（用于打印）
func (l *Logger) Output() string {
	return l.buf.String()
}
