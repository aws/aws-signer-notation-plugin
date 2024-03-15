//  Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License"). You may
//  not use this file except in compliance with the License. A copy of the
//  License is located at
//
// 	http://aws.amazon.com/apache2.0
//
//  or in the "license" file accompanying this file. This file is distributed
//  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language governing
//  permissions and limitations under the License.

package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugLogger(t *testing.T) {
	t.Run("DebugLoggerNotEnabled", func(t *testing.T) {
		tempDir, _ := os.MkdirTemp("", "tempDir")
		defer os.RemoveAll(tempDir)

		userConfigDir = func() (string, error) {
			return tempDir, nil
		}
		logFilePath := filepath.Join(tempDir, "notation-aws-signer", "plugin.log")
		GetLogger(context.TODO()).Debug("This is a Debug log!")
		_, err := os.ReadFile(logFilePath)
		if err == nil {
			t.Fatalf("debug log file found at " + logFilePath)
		}
	})

	t.Run("DebugLoggerEnabled", func(t *testing.T) {
		logger, logFilePath := setupTestLogger(t)
		msg := "This is a Debug log!"
		logger.Debug(msg)
		validateLogEntry(logFilePath, "[DEBUG] "+msg, t)
	})

	t.Run("UnableToCreateFile", func(t *testing.T) {
		userConfigDir = func() (string, error) {
			return "", fmt.Errorf("expected error thrown")
		}
		_, err := New()
		assert.Error(t, err, "expected error not found")
	})
}

func TestDebugLogger_UpdateContext(t *testing.T) {
	ctx := context.TODO()
	logger, _ := setupTestLogger(t)
	ctx = logger.UpdateContext(ctx)
	ctxLogger := GetLogger(ctx)
	assert.Equal(t, logger, ctxLogger, "debugLogger in context should match the original logger")
}

func TestDebugLogger_Info(t *testing.T) {
	logger, logFilePath := setupTestLogger(t)
	msg := "This is a Info log!"
	logger.Info(msg)
	validateLogEntry(logFilePath, "[INFO] "+msg, t)

	logger.Infoln(msg)
	validateLogEntry(logFilePath, "[INFO] "+msg+"\n", t)
}

func TestDebugLogger_Warn(t *testing.T) {
	logger, logFilePath := setupTestLogger(t)
	msg := "This is a Warn log!"
	logger.Warn(msg)
	validateLogEntry(logFilePath, "[WARN] "+msg, t)

	logger.Warnln(msg)
	validateLogEntry(logFilePath, "[WARN] "+msg+"\n", t)
}

func TestDebugLogger_Debug(t *testing.T) {
	logger, logFilePath := setupTestLogger(t)
	msg := "This is a Debug log!"
	logger.Debug(msg)
	validateLogEntry(logFilePath, "[DEBUG] "+msg, t)

	logger.Debugln(msg)
	validateLogEntry(logFilePath, "[DEBUG] "+msg+"\n", t)
}

func TestDebugLogger_Error(t *testing.T) {
	logger, logFilePath := setupTestLogger(t)
	msg := "This is a Debug log!"
	logger.Error(msg)
	validateLogEntry(logFilePath, "[ERROR] "+msg, t)

	logger.Errorln(msg)
	validateLogEntry(logFilePath, "[ERROR] "+msg+"\n", t)
}

func setupTestLogger(t *testing.T) (*debugLogger, string) {
	userConfigDir = func() (string, error) {
		return os.TempDir(), nil
	}
	expectedLogFilePath := filepath.Join(os.TempDir(), "notation-aws-signer", "plugin.log")
	logger, _ := New()
	t.Cleanup(func() {
		logger.Close()
		os.RemoveAll(expectedLogFilePath)
	})

	return logger, expectedLogFilePath
}

func validateLogEntry(logfilePath, expectedLogLine string, t *testing.T) {
	logs, err := os.ReadFile(logfilePath)
	if err != nil {
		t.Fatalf("Log file not found at %s. Error: %v", logfilePath, err)
	}
	if !strings.HasSuffix(string(logs), expectedLogLine) {
		t.Errorf("Expected log message '%s' not found in the log file '%s'", expectedLogLine, string(logs))
	}
}
