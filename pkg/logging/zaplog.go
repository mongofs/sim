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
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)



type OutPut int

const (
	OutPutFile OutPut = iota + 1
	OutPutStout
)

type log struct {
	option *Option
	*zap.Logger
}

const (
	DefaultLogName  = "sim"
	DefaultLogPath  = "./log"
	DefaultLogLevel = InfoLevel // debug、info、warn、error、panic、fatal
)

var (
	defaultOption = &Option{
		logName: DefaultLogName,
		logPath: DefaultLogPath,
		level:   DefaultLogLevel,
	}
	Log = &log{
		option: defaultOption,
		Logger: nil,
	}
)

type Level int

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

type Option struct {
	logName string
	logPath string
	level   Level
}

type OptionFunc func(option *Option)

func SetLogName(logName string) OptionFunc {
	return func(o *Option) {
		o.logName = logName
	}
}
func SetLogPath(logPath string) OptionFunc {
	return func(o *Option) {
		o.logName = logPath
	}
}
func SetLevel(level Level) OptionFunc {
	return func(o *Option) {
		o.level = level
	}
}

func init(){
	core := getCores(OutPutStout, Log.option.logName)
	caller := zap.AddCaller()
	development := zap.Development()
	zlog := zap.New(core, caller, development)
	Log.Logger = zlog
}

func InitZapLogger(bug bool, ops ...OptionFunc) *log {
	var out OutPut
	if !bug {
		out = OutPutFile
	} else {
		out = OutPutStout
	}
	for _, v := range ops {
		v(Log.option)
	}
	core := getCores(out, Log.option.logName)
	caller := zap.AddCaller()
	development := zap.Development()
	zlog := zap.New(core, caller, development)
	Log.Logger = zlog
	return Log
}
func getZapEncoder() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
}
func getWriter(logPath, logfile string) lumberjack.Logger {
	if logPath == "" {
		logPath = "./log"
	}
	today := time.Now().Format("20060102")
	filename := fmt.Sprintf("%s/%s/%s", logPath, today, logfile)
	return lumberjack.Logger{
		Filename:   filename, // 日志文件路径
		MaxSize:    128,      // 每个日志文件保存的最大尺寸 单位：M  128
		MaxBackups: 30,       // 日志文件最多保存多少个备份 30
		MaxAge:     7,        // 文件最多保存多少天 7
		Compress:   true,     // 是否压缩
	}
}
func getCores(output OutPut, serverName string) zapcore.Core {
	cores := []zapcore.Core{}
	encoderConfig := getZapEncoder()

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl == zapcore.InfoLevel })
	waringLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl == zapcore.WarnLevel })
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl == zapcore.ErrorLevel })
	switch output {
	case OutPutStout:
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), infoLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), waringLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), errorLevel))
	case OutPutFile:
		// 获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
		infoWriter := getWriter("", serverName+"_info.log")
		waringWriter := getWriter("", serverName+"_waring.log")
		errorWriter := getWriter("", serverName+"_error.log")
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&infoWriter)), infoLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&waringWriter)), waringLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&errorWriter)), errorLevel))
	}
	return zapcore.NewTee(cores...)
}
