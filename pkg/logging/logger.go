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
	"errors"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)


var (
	flushLogs           func() error
	defaultLogger       Logger
	defaultLoggingLevel Level
)

// Level is the alias of zapcore.Level.
type Level = zapcore.Level

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
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

	// Initializes the inside default logger of sim.
	fileName := os.Getenv("LOGGINH_FILE")
	if len(fileName) == 0 {
		fileName = "./log/steven.log"
	}
	if len(fileName) > 0 {
		var err error
		defaultLogger, flushLogs, err = CreateLoggerAsLocalFile(fileName, defaultLoggingLevel)
		if err != nil {
			panic("invalid LOGGINH_FILE, " + err.Error())
		}
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(defaultLoggingLevel)
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
		zapLogger, _ := cfg.Build()
		defaultLogger = zapLogger.Sugar()
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// GetDefaultLogger returns the default logger.
func GetDefaultLogger() Logger {
	return defaultLogger
}

// LogLevel tells what the default logging level is.
func LogLevel() string {
	return defaultLoggingLevel.String()
}

// CreateLoggerAsLocalFile setups the logger by local file path.
func CreateLoggerAsLocalFile(localFilePath string, logLevel Level) (logger Logger, flush func() error, err error) {
	if len(localFilePath) == 0 {
		return nil, nil, errors.New("invalid local logger path")
	}

	// lumberjack.Logger is already safe for concurrent use, so we don't need to lock it.
	lumberJackLogger := &lumberjack.Logger{
		Filename:   localFilePath,
		MaxSize:    100, // megabytes
		MaxBackups: 2,
		MaxAge:     15, // days
	}

	encoder := getEncoder()
	ws := zapcore.AddSync(lumberJackLogger)
	zapcore.Lock(ws)

	levelEnabler := zap.LevelEnablerFunc(func(level Level) bool {
		return level >= logLevel
	})
	core := zapcore.NewCore(encoder, ws, levelEnabler)
	zapLogger := zap.New(core, zap.AddCaller())
	logger = zapLogger.Sugar()
	flush = zapLogger.Sync
	return
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


