/*
 * Copyright 2022 steven
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *    http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strconv"
)

var (
	flushLogs           func() error
	defaultLogger       Logger
	defaultLoggingLevel Level
)

// Level is the alias of zapcore.Level.
type Level = zapcore.Level

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)


func init() {
	lvl := os.Getenv("LOGGING_LEVEL")
	if len(lvl) > 0 {
		loggingLevel, err := strconv.ParseInt(lvl, 10, 8)
		if err != nil {
			panic("invalid LOGGING_LEVEL Setting, " + err.Error())
		}
		defaultLoggingLevel = Level(loggingLevel)
	}

	// fileName 不要后缀
	fileName := os.Getenv("LOGGINH_FILE")
	if len(fileName) == 0 {
		fileName = "sim"
	}

	core := getCores(OutputStdout, fileName)
	caller := zap.AddCaller()
	development := zap.Development()
	zaplog := zap.New(core, caller, development)
	defaultLogger= &log{zaplog}
}

// FlushLogPath
func FlushLogPath(logPath,logFile string){
	core := getCores(OutputStdout, logFile)
	caller := zap.AddCaller()
	development := zap.Development()
	zaplog := zap.New(core, caller, development)
	defaultLogger= &log{zaplog}
}



// GetDefaultLogger returns the default logger.
func GetDefaultLogger() Logger {
	return defaultLogger
}


// Logger is used for logging formatted messages.
type Logger interface {
	// Debugf logs messages at DEBUG level.
	Debugf(format string, args ...interface{})
	// Infof logs messages at INFO level.
	Infof(format string, args ...interface{})
	// Warnf logs messages at WARN level.
	Warnf(format string, args ...interface{})
	// Errorf logs messages at ERROR level.
	Errorf(format string, args ...interface{})
	// Fatalf logs messages at FATAL level.
	Fatalf(format string, args ...interface{})
}

// Cleanup does something windup for logger, like closing, flushing, etc.
func Cleanup() {
	if flushLogs != nil {
		_ = flushLogs()
	}
}

// Error prints err if it's not nil.
func Error(err error) {
	if err != nil {
		defaultLogger.Errorf("error occurs during runtime, %v", err)
	}
}

// Debugf logs messages at DEBUG level.
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Infof logs messages at INFO level.
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warnf logs messages at WARN level.
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Errorf logs messages at ERROR level.
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatalf logs messages at FATAL level.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}



