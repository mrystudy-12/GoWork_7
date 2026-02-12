package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger 日志结构体
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewLogger 创建新的日志器
func NewLogger(logDir string) *Logger {
	// 确保日志目录存在
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 创建日志文件
	logFile := filepath.Join(logDir, fmt.Sprintf("%s.log", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}

	// 创建日志器
	return &Logger{
		infoLogger:  log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info 记录信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Output(2, fmt.Sprintf(format, v...))
	fmt.Printf("[INFO] %s\n", fmt.Sprintf(format, v...))
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Output(2, fmt.Sprintf(format, v...))
	fmt.Printf("[ERROR] %s\n", fmt.Sprintf(format, v...))
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Output(2, fmt.Sprintf(format, v...))
	fmt.Printf("[DEBUG] %s\n", fmt.Sprintf(format, v...))
}

// 全局日志器
var (
	// AuthLogger 认证相关日志
	AuthLogger = NewLogger("logs/auth")
	// SystemLogger 系统相关日志
	SystemLogger = NewLogger("logs/system")
	// UserLogger 用户相关日志
	UserLogger = NewLogger("logs/user")
)

// InitLoggers 初始化日志记录器
func InitLoggers() error {
	// 日志记录器在包初始化时已经创建，这里可以添加额外的初始化逻辑
	return nil
}
