// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logCoreFactoryFunc func() zapcore.Core
)

func SetLogCoreFactory(f func() zapcore.Core) {
	logCoreFactoryFunc = f
}

func (l *Logger) Init() {
	var zapLogger *zap.Logger
	if logCoreFactoryFunc != nil {
		zapLogger = zap.New(logCoreFactoryFunc())
	} else {
		zapLogger = zap.New(zapcore.NewTee(
			zapcore.NewCore(l.getEncoder(l.encoder.LevelKey), l.getLogWriter(l.infoLogPath), infoLevelEnabler),
			zapcore.NewCore(l.getEncoder(l.encoder.LevelKey), l.getLogWriter(l.errorLogPath), errorLevelEnabler),
			zapcore.NewCore(l.getEncoder(""), l.getLogWriter(l.accessLogPath), accessLevelEnabler),
		), zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(errorLevelEnabler))
	}
	l.sugaredLogger = zapLogger.Sugar()
	l.logger        = zapLogger
}
