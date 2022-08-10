package log

import (
	"fmt"
	"strings"
)

// NewSubLogger allows for a new sub logger to be registered.
func NewSubLogger(name string) (*SubLogger, error) {
	if name == "" {
		return nil, errEmptyLoggerName
	}
	name = strings.ToUpper(name)
	mu.Lock()
	defer mu.Unlock()
	if _, ok := SubLoggers[name]; ok {
		return nil, fmt.Errorf("'%v' %w", name, ErrSubLoggerAlreadyRegistered)
	}
	return registerNewSubLogger(name), nil
}

// SetOutput overrides the default output with a new writer
func (sl *SubLogger) setOutput(o *multiWriterHolder) {
	sl.output = o
}

// SetLevels overrides the default levels with new levels; levelception
func (sl *SubLogger) setLevels(newLevels Levels) {
	sl.levels = newLevels
}

func (sl *SubLogger) getFields() *logFields {
	if sl == nil || globalLogConfig == nil || globalLogConfig.Enabled == nil || !*globalLogConfig.Enabled {
		return nil
	}

	fields := logFieldsPool.Get().(*logFields) // nolint:forcetypeassert // Not necessary from a pool
	fields.info = sl.levels.Info
	fields.warn = sl.levels.Warn
	fields.debug = sl.levels.Debug
	fields.error = sl.levels.Error
	fields.name = sl.name
	fields.output = sl.output
	return fields
}
