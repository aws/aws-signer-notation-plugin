// Package logger provides logging functionality.
package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-signer-notation-plugin/internal/version"
)

type contextKey int

const logContextKey contextKey = iota

var userConfigDir = os.UserConfigDir // for unit test
var discardLogger = &debugLogger{}

type debugLogger struct {
	file *os.File
}

// New creates a new debugLogger instance
func New() (*debugLogger, error) {
	cfgDir, err := userConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(cfgDir, "notation-aws-signer", "plugin.log")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}

	dl := &debugLogger{
		file: file,
	}

	dl.Debugln("-----------------------------------------------------------------------------------------")
	dl.Debugf("Logs from execution of AWS signer plugin version: %s\n", version.GetVersion())
	return dl, nil
}

// Close closes the logger and associated resources
func (l *debugLogger) Close() {
	if l.file != nil {
		err := l.file.Close()
		if err != nil {
			l.Errorf("error while closing the log file", err)
		}
	}
}

// GetLogger returns the logger instance
func GetLogger(ctx context.Context) *debugLogger {
	debugLogger, ok := ctx.Value(logContextKey).(*debugLogger)
	if !ok {
		return discardLogger
	}
	return debugLogger
}

// UpdateContext returns context with the logger entry
func (l *debugLogger) UpdateContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, logContextKey, l)
}

// IsDebug returns true if Debug log is enabled
func (l *debugLogger) IsDebug() bool {
	return l.file != nil
}

func (l *debugLogger) Debug(args ...interface{}) {
	l.log("DEBUG", args...)
}

func (l *debugLogger) Debugf(format string, args ...interface{}) {
	l.logf("DEBUG", format, args...)
}

func (l *debugLogger) Debugln(args ...interface{}) {
	l.logln("DEBUG", args...)
}

func (l *debugLogger) Info(args ...interface{}) {
	l.log("INFO", args...)
}

func (l *debugLogger) Infof(format string, args ...interface{}) {
	l.logf("INFO", format, args...)
}

func (l *debugLogger) Infoln(args ...interface{}) {
	l.logln("INFO", args...)
}

func (l *debugLogger) Warn(args ...interface{}) {
	l.log("WARN", args...)
}

func (l *debugLogger) Warnf(format string, args ...interface{}) {
	l.logf("WARN", format, args...)
}

func (l *debugLogger) Warnln(args ...interface{}) {
	l.logln("WARN", args...)
}

func (l *debugLogger) Error(args ...interface{}) {
	l.log("ERROR", args...)
}

func (l *debugLogger) Errorf(format string, args ...interface{}) {
	l.logf("ERROR", format, args...)
}

func (l *debugLogger) Errorln(args ...interface{}) {
	l.logln("ERROR", args...)
}

func (l *debugLogger) logf(levelPrefix, format string, args ...interface{}) {
	if l.file != nil {
		_, _ = fmt.Fprintf(l.file, "%s [%s] "+format, append([]interface{}{time.Now().Format(time.RFC3339Nano), levelPrefix}, args...)...)
	}
}

func (l *debugLogger) log(levelPrefix string, args ...interface{}) {
	l.logf(levelPrefix, "%v", args...)
}

func (l *debugLogger) logln(levelPrefix string, args ...interface{}) {
	l.logf(levelPrefix, "%v\n", args...)
}
