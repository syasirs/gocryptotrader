package logger

import (
	"io"
	"sync"
)

const timestampFormat = " 02/01/2006 15:04:05 "

const spacer = "|"

type Config struct {
	Enabled          *bool            `json:"enabled"`
	AdvancedSettings advancedSettings `json:"advancedSettings"`
	SubLoggers       []subLoggers     `json:"subloggers"`
}

type headers struct {
	Info  string `json:"info"`
	Warn  string `json:"warn"`
	Debug string `json:"debug"`
	Error string `json:"error"`
}

type advancedSettings struct {
	Spacer          string  `json:"spacer"`
	TimeStampFormat string  `json:"timeStampFormat"`
	Headers         headers `json:"headers"`
}

type subLoggers struct {
	Name   string `json:"name"`
	Level  string `json:"level"`
	Output string `json:"output"`
}

type Logger struct {
	Timestamp                                        string
	InfoHeader, ErrorHeader, DebugHeader, WarnHeader string
	Spacer                                           string
}

type subLogger struct {
	Info, Debug, Warn, Error bool
	output                   io.Writer
}

type LogEvent struct {
	data   []byte
	output io.Writer
}

type multiWriter struct {
	writers []io.Writer
	mu      sync.Mutex
}

var (
	logger           = &Logger{}
	GlobalLogConfig  = &Config{}
	subsystemLoggers = map[string]subLogger{}
	eventPool        = &sync.Pool{
		New: func() interface{} {
			return &LogEvent{
				data: make([]byte, 0, 80),
			}
		},
	}
	LogPath string
)

const (
	LogGlobal = "log"

	SubSystemConnMgr = "connmgr"
	SubSystemCommMgr = "commmgr"
	SubSystemConfMgr = "confmgr"
	SubSystemOrdrMgr = "ordrmgr"
	SubSystemPortMgr = "portmgr"
	SubSystemSyncMgr = "syncmsr"
	SubSystemTimeMgr = "timemgr"
	SubSystemWsocMgr = "wsocmgr"

	SubSystemExchSys = "exchsys"
	SubSystemGrpcSys = "grpcsys"
	SubSystemRestSys = "restsys"

	SubSystemTicker    = "ticker"
	SubSystemOrderBook = "orderbook"
)
