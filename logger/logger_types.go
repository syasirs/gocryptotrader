package logger

import (
	"io"
	"sync"
)

const (
	timestampFormat = " 02/01/2006 15:04:05 "
	spacer          = "|"
	MaxFileSize int64 = 100
)

var (
	logger = &Logger{}
	// FileLoggingConfiguredCorrectly flag set during config check if file logging meets requirements
	FileLoggingConfiguredCorrectly bool
	// GlobalLogConfig holds global configuration options for logger
	GlobalLogConfig = &Config{}
	// GlobalLogFile hold global configuration options for file logger
	GlobalLogFile = &Rotate{}

	eventPool = &sync.Pool{
		New: func() interface{} {
			return &LogEvent{
				data: make([]byte, 0, 80),
			}
		},
	}

	// LogPath system path to store log files in
	LogPath string
)

// Config holds configuration settings loaded from bot config
type Config struct {
	Enabled bool `json:"enabled"`
	SubLoggerConfig
	LoggerFileConfig *loggerFileConfig `json:"fileSettings,omitempty"`
	AdvancedSettings advancedSettings  `json:"advancedSettings"`
	SubLoggers       []SubLoggerConfig `json:"subloggers,omitempty"`
}

type advancedSettings struct {
	Spacer          string  `json:"spacer"`
	TimeStampFormat string  `json:"timeStampFormat"`
	Headers         headers `json:"headers"`
}

type headers struct {
	Info  string `json:"info"`
	Warn  string `json:"warn"`
	Debug string `json:"debug"`
	Error string `json:"error"`
}

// SubLoggerConfig holds sub logger configuration settings loaded from bot config
type SubLoggerConfig struct {
	Name   string `json:"name,omitempty"`
	Level  string `json:"level"`
	Output string `json:"output"`
}

type loggerFileConfig struct {
	FileName string `json:"filename,omitempty"`
	Rotate   *bool  `json:"rotate,omitempty"`
	MaxSize  int64  `json:"maxsize,omitempty"`
}

// Logger each instance of logger settings
type Logger struct {
	Timestamp                                        string
	InfoHeader, ErrorHeader, DebugHeader, WarnHeader string
	Spacer                                           string
}

// Levels flags for each sub logger type
type Levels struct {
	Info, Debug, Warn, Error bool
}

type subLogger struct {
	name string
	Levels
	output io.Writer
}

// LogEvent holds the data sent to the log and which multiwriter to send to
type LogEvent struct {
	data   []byte
	output io.Writer
}

type multiWriter struct {
	writers []io.Writer
	mu      sync.RWMutex
}
