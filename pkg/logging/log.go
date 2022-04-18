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

type log struct {
	*zap.Logger
}

func (l log) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format,args...))
}

func (l log) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format,args...))
}

func (l log) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format,args...))
}

func (l log) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format,args...))
}

func (l log) Fatalf(format string, args ...interface{}) {
	l.Fatal(fmt.Sprintf(format,args...))
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

func getWriter(logPath,logfile string) lumberjack.Logger {
	if logPath == "" {
		logPath = "./log"
	}
	today := time.Now().Format("20060102")
	filename := fmt.Sprintf("%s/%s/%s", logPath,today, logfile)
	return lumberjack.Logger{
		Filename:   filename, // 日志文件路径
		MaxSize:    128,      // 每个日志文件保存的最大尺寸 单位：M  128
		MaxBackups: 30,       // 日志文件最多保存多少个备份 30
		MaxAge:     7,        // 文件最多保存多少天 7
		Compress:   true,     // 是否压缩
	}
}



func getCores(output output, serverName string) zapcore.Core {
	cores := []zapcore.Core{}
	encoderConfig := getZapEncoder()

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl == zapcore.InfoLevel })
	waringLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl == zapcore.WarnLevel })
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl == zapcore.ErrorLevel })
	switch output {
	case OutputStdout:
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), infoLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), waringLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), errorLevel))
	case OutputFile:
		// 获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
		infoWriter := getWriter("",serverName + "_info.log")
		waringWriter := getWriter("",serverName + "_waring.log")
		errorWriter := getWriter("",serverName + "_error.log")
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&infoWriter)), infoLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&waringWriter)), waringLevel))
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&errorWriter)), errorLevel))
	}
	return zapcore.NewTee(cores...)
}
