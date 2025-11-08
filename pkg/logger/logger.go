package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	SetupLoggerWithFile()
}

func SetupLoggerWithFile() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create logs directory structure with date
	dateStr := time.Now().Format("2006-01-02")
	logsDir := filepath.Join("logs", dateStr)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("Failed to create logs directory")
	}

	// Create separate log files in date directory
	apiLogFile := createLogFile(logsDir, "api.log")
	errorLogFile := createLogFile(logsDir, "error.log")
	accessLogFile := createLogFile(logsDir, "access.log")
	systemLogFile := createLogFile(logsDir, "system.log")

	// Console writer for development
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Setup main logger (API logs)
	apiMulti := zerolog.MultiLevelWriter(consoleWriter, apiLogFile)
	log.Logger = zerolog.New(apiMulti).With().
		Timestamp().
		Str("type", "api").
		Logger()

	// Setup global loggers for different types
	setupGlobalLoggers(apiLogFile, errorLogFile, accessLogFile, systemLogFile, consoleWriter)
}

// Global loggers for different log types
var (
	APILogger    zerolog.Logger
	ErrorLogger  zerolog.Logger
	AccessLogger zerolog.Logger
	SystemLogger zerolog.Logger
)

func setupGlobalLoggers(apiFile, errorFile, accessFile, systemFile *os.File, console zerolog.ConsoleWriter) {
	APILogger = zerolog.New(zerolog.MultiLevelWriter(console, apiFile)).With().
		Timestamp().
		Str("type", "api").
		Logger()

	ErrorLogger = zerolog.New(zerolog.MultiLevelWriter(console, errorFile)).With().
		Timestamp().
		Str("type", "error").
		Logger()

	AccessLogger = zerolog.New(accessFile).With().
		Timestamp().
		Str("type", "access").
		Logger()

	SystemLogger = zerolog.New(systemFile).With().
		Timestamp().
		Str("type", "system").
		Logger()
}

func createLogFile(dir, filename string) *os.File {
	filePath := filepath.Join(dir, filename)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Str("file", filePath).Msg("Failed to create log file")
	}
	return file
}
