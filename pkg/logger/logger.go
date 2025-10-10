package logger

import (
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

// Init initializes the application loggers
func Init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info logs info messages
func Info(v ...interface{}) {
	InfoLogger.Println(v...)
}

// Infof logs formatted info messages
func Infof(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// Error logs error messages
func Error(v ...interface{}) {
	ErrorLogger.Println(v...)
}

// Errorf logs formatted error messages
func Errorf(format string, v ...interface{}) {
	ErrorLogger.Printf(format, v...)
}

// Debug logs debug messages
func Debug(v ...interface{}) {
	DebugLogger.Println(v...)
}

// Debugf logs formatted debug messages
func Debugf(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}

// Fatal logs fatal error and exits
func Fatal(v ...interface{}) {
	ErrorLogger.Fatal(v...)
}

// Fatalf logs formatted fatal error and exits
func Fatalf(format string, v ...interface{}) {
	ErrorLogger.Fatalf(format, v...)
}
