package fyersgosdk

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type FyersLogger struct {
	service    string
	level      string
	stackLevel int
	logger     *log.Logger
	logFile    *os.File
	location   string
}

func NewFyersLogger(service, level string, stackLevel int, logPath string) *FyersLogger {
	if stackLevel == 0 {
		stackLevel = 4
	}

	if logPath != "" {
		err := os.MkdirAll(logPath, 0755)
		if err != nil {
			log.Printf("Failed to create log directory: %v", err)
		}
	}

	var logFile *os.File
	var err error
	if logPath != "" {
		var logFileName string
		if strings.Contains(service, "Order") {
			logFileName = "fyersOrderSocket.log"
		} else {
			logFileName = "fyersDataSocket.log"
		}

		logFilePath := filepath.Join(logPath, logFileName)
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open log file: %v", err)
		}
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	if logFile != nil {
		logger.SetOutput(logFile)
	}

	return &FyersLogger{
		service:    service,
		level:      level,
		stackLevel: stackLevel,
		logger:     logger,
		logFile:    logFile,
	}
}

func (f *FyersLogger) getLocation() string {
	if f.location != "" {
		return f.location
	}

	pc, file, line, ok := runtime.Caller(f.stackLevel)
	if !ok {
		return "[unknown:0] unknown"
	}

	funcName := runtime.FuncForPC(pc).Name()
	parts := strings.Split(funcName, ".")
	funcName = parts[len(parts)-1]

	parts = strings.Split(file, "/")
	fileName := parts[len(parts)-1]

	f.location = fmt.Sprintf("[%s:%d] %s", funcName, line, fileName)
	return f.location
}

func (f *FyersLogger) formatMessage(level, msg string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	location := f.getLocation()
	return fmt.Sprintf("[%s] [%s] [%s] %s - %s", timestamp, f.service, level, location, msg)
}

func (f *FyersLogger) Error(msg interface{}) {
	message := fmt.Sprintf("%v", msg)
	formattedMsg := f.formatMessage("ERROR", message)
	f.logger.Println(formattedMsg)
}

func (f *FyersLogger) Info(msg interface{}) {
	message := fmt.Sprintf("%v", msg)
	formattedMsg := f.formatMessage("INFO", message)
	f.logger.Println(formattedMsg)
}

func (f *FyersLogger) Debug(msg interface{}) {
	if f.level == "DEBUG" {
		message := fmt.Sprintf("%v", msg)
		formattedMsg := f.formatMessage("DEBUG", message)
		f.logger.Println(formattedMsg)
	}
}

func (f *FyersLogger) Exception(msg interface{}) {
	message := fmt.Sprintf("%v", msg)
	formattedMsg := f.formatMessage("EXCEPTION", message)
	f.logger.Println(formattedMsg)
}

func (f *FyersLogger) Close() {
	if f.logFile != nil {
		f.logFile.Close()
	}
}
