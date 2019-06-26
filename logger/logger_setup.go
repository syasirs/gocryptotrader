package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func getWriters(s *SubLoggerConfig) io.Writer {
	mw := MultiWriter()
	m := mw.(*multiWriter)

	outputWriters := strings.Split(s.Output, "|")
	for x := range outputWriters {
		switch outputWriters[x] {
		case "stdout":
			m.Add(os.Stdout)
		case "stderr":
			m.Add(os.Stderr)
		case "file":
			if GlobalLogFile != nil {
				m.Add(GlobalLogFile)
			}
		case "ownfile":
			temp, err := createFileHandle(s.Name, false)
			if err != nil {
				fmt.Printf("File handle error %v", err)
			}
			m.Add(temp)
		default:
			m.Add(ioutil.Discard)
		}
	}
	return m
}

// GenDefaultSettings return struct with known sane/working logger settings
func GenDefaultSettings() (log Config) {
	t := func(t bool) *bool { return &t }(true)
	log = Config{
		Enabled: t,
		SubLoggerConfig: SubLoggerConfig{
			Level:  "INFO|DEBUG|WARN|ERROR",
			Output: "stdout",
		},
		AdvancedSettings: advancedSettings{
			Spacer:          " | ",
			TimeStampFormat: timestampFormat,
			Headers: headers{
				Info:  "[INFO]",
				Warn:  "[WARN]",
				Debug: "[DEBUG]",
				Error: "[ERROR]",
			},
		},
	}
	return
}

func configureSubLogger(logger, levels string, output io.Writer) {
	found, logPtr := validSubLogger(logger)
	if !found {
		return
	}

	logPtr.output = output

	logPtr.levels = splitLevel(levels)
	subLoggers[logger] = logPtr
}

// SetupSubLoggers configure all sub loggers with provided configuration values
func SetupSubLoggers(s []SubLoggerConfig) {
	for x := range s {
		output := getWriters(&s[x])
		configureSubLogger(s[x].Name, s[x].Level, output)
	}
}

// SetupGlobalLogger() setup the global loggers with the default global config values
func SetupGlobalLogger() {
	var err error
	GlobalLogFile, err = createFileHandle("", true)
	if err != nil {
		fmt.Printf("Failed to open log file global file logging unavailable %v", err)
	}
	for x := range subLoggers {
		subLoggers[x].levels = splitLevel(GlobalLogConfig.Level)
		subLoggers[x].output = getWriters(&GlobalLogConfig.SubLoggerConfig)
	}

	logger = newLogger(GlobalLogConfig)
}

func splitLevel(level string) (l levels) {
	enabledLevels := strings.Split(level, "|")
	for x := range enabledLevels {
		switch level := enabledLevels[x]; level {
		case "DEBUG":
			l.Debug = true
		case "INFO":
			l.Info = true
		case "WARN":
			l.Warn = true
		case "ERROR":
			l.Error = true
		}
	}
	return
}

func registerNewSubLogger(logger string) *subLogger {
	temp := subLogger{
		name:   logger,
		output: os.Stdout,
	}

	temp.levels = splitLevel("INFO|WARN|DEBUG|ERROR")
	subLoggers[logger] = &temp

	return &temp
}

// register all loggers at package init()
func init() {
	Global = registerNewSubLogger("log")

	ConnectionMgr = registerNewSubLogger("connection")
	CommunicationMgr = registerNewSubLogger("comms")
	ConfigMgr = registerNewSubLogger("config")
	OrderMgr = registerNewSubLogger("order")
	PortfolioMgr = registerNewSubLogger("portfolio")
	SyncMgr = registerNewSubLogger("sync")
	TimeMgr = registerNewSubLogger("timekeeper")
	WebsocketMgr = registerNewSubLogger("websocket")
	EventMgr = registerNewSubLogger("event")

	ExchangeSys = registerNewSubLogger("exchange")
	GRPCSys = registerNewSubLogger("grpc")
	RESTSys = registerNewSubLogger("rest")

	Ticker = registerNewSubLogger("ticker")
	OrderBook = registerNewSubLogger("orderbook")
}
