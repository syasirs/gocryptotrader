package log

import (
	"fmt"
	"log"
)

// Info takes a pointer subLogger struct and string sends to newLogEvent
func Info(sl *SubLogger, data string) {
	fields := sl.getFields()
	if fields == nil || !fields.info {
		return
	}

	displayError(fields.logger.newLogEvent(data,
		fields.logger.InfoHeader,
		fields.name,
		fields.output))
}

// Infoln takes a pointer subLogger struct and interface sends to newLogEvent
func Infoln(sl *SubLogger, v ...interface{}) {
	fields := sl.getFields()
	if fields == nil || !fields.info {
		return
	}
	displayError(fields.logger.newLogEvent(fmt.Sprintln(v...),
		fields.logger.InfoHeader,
		fields.name,
		fields.output))
}

// Infof takes a pointer subLogger struct, string & interface formats and sends to Info()
func Infof(sl *SubLogger, data string, v ...interface{}) {
	Info(sl, fmt.Sprintf(data, v...))
}

// Debug takes a pointer subLogger struct and string sends to multiwriter
func Debug(sl *SubLogger, data string) {
	fields := sl.getFields()
	if fields == nil || !fields.debug {
		return
	}
	displayError(fields.logger.newLogEvent(data,
		fields.logger.DebugHeader,
		fields.name,
		fields.output))
}

// Debugln  takes a pointer subLogger struct, string and interface sends to newLogEvent
func Debugln(sl *SubLogger, v ...interface{}) {
	fields := sl.getFields()
	if fields == nil || !fields.debug {
		return
	}

	displayError(fields.logger.newLogEvent(fmt.Sprintln(v...),
		fields.logger.DebugHeader,
		fields.name,
		fields.output))
}

// Debugf takes a pointer subLogger struct, string & interface formats and sends to Info()
func Debugf(sl *SubLogger, data string, v ...interface{}) {
	Debug(sl, fmt.Sprintf(data, v...))
}

// Warn takes a pointer subLogger struct & string  and sends to newLogEvent()
func Warn(sl *SubLogger, data string) {
	fields := sl.getFields()
	if fields == nil || !fields.warn {
		return
	}
	displayError(fields.logger.newLogEvent(data,
		fields.logger.WarnHeader,
		fields.name,
		fields.output))
}

// Warnln takes a pointer subLogger struct & interface formats and sends to newLogEvent()
func Warnln(sl *SubLogger, v ...interface{}) {
	fields := sl.getFields()
	if fields == nil || !fields.warn {
		return
	}
	displayError(fields.logger.newLogEvent(fmt.Sprintln(v...),
		fields.logger.WarnHeader,
		fields.name,
		fields.output))
}

// Warnf takes a pointer subLogger struct, string & interface formats and sends to Warn()
func Warnf(sl *SubLogger, data string, v ...interface{}) {
	Warn(sl, fmt.Sprintf(data, v...))
}

// Error takes a pointer subLogger struct & interface formats and sends to newLogEvent()
func Error(sl *SubLogger, data ...interface{}) {
	fields := sl.getFields()
	if fields == nil || !fields.error {
		return
	}
	displayError(fields.logger.newLogEvent(fmt.Sprint(data...),
		fields.logger.ErrorHeader,
		fields.name,
		fields.output))
}

// Errorln takes a pointer subLogger struct, string & interface formats and sends to newLogEvent()
func Errorln(sl *SubLogger, v ...interface{}) {
	fields := sl.getFields()
	if fields == nil || !fields.error {
		return
	}
	displayError(fields.logger.newLogEvent(fmt.Sprintln(v...),
		fields.logger.ErrorHeader,
		fields.name,
		fields.output))
}

// Errorf takes a pointer subLogger struct, string & interface formats and sends to Debug()
func Errorf(sl *SubLogger, data string, v ...interface{}) {
	Error(sl, fmt.Sprintf(data, v...))
}

func displayError(err error) {
	if err != nil {
		log.Printf("Logger write error: %v\n", err)
	}
}

func enabled() bool {
	return GlobalLogConfig != nil &&
		GlobalLogConfig.Enabled != nil &&
		*GlobalLogConfig.Enabled
}

func (sl *SubLogger) getFields() *logFields {
	RWM.RLock()
	defer RWM.RUnlock()

	if !enabled() && sl == nil {
		return nil
	}

	return &logFields{
		info:   sl.Info,
		warn:   sl.Warn,
		debug:  sl.Debug,
		error:  sl.Error,
		name:   sl.name,
		output: sl.output,
		logger: logger,
	}
}
