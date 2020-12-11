// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package logger

import (
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

var logLayout = "2006-01-02T15:04:05.000Z0700"

var _logEmptyStrSlice = [...][]byte{
	[]byte("       -"),
	[]byte("      -"),
	[]byte("     -"),
	[]byte("    -"),
	[]byte("   -"),
	[]byte("  -"),
	[]byte(" -"),
	[]byte(" -"),
	[]byte(" -"),
}

func SetLogLayout(layout string) {
	logLayout = layout
}

func adminTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	encodeTimeLayout(t, logLayout, enc)
}

func adminLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var sb strings.Builder
	sb.Grow(10)
	sb.Write([]byte("- "))
	s := l.CapitalString()
	sb.WriteString(s)
	sb.Write(_logEmptyStrSlice[len(s)])
	enc.AppendString(sb.String())
}

func (l *Logger) getEncoder(levelKey string) zapcore.Encoder {

	var (
		//timeEncoder     = new(zapcore.TimeEncoder)
		durationEncoder = new(zapcore.DurationEncoder)
		callerEncoder   = new(zapcore.CallerEncoder)
		nameEncoder     = new(zapcore.NameEncoder)
		//levelEncoder    = new(zapcore.LevelEncoder)
	)

	//_ = timeEncoder.UnmarshalText([]byte(l.encoder.Time))
	_ = durationEncoder.UnmarshalText([]byte(l.encoder.Duration))
	_ = callerEncoder.UnmarshalText([]byte(l.encoder.Caller))
	_ = nameEncoder.UnmarshalText([]byte("full"))
	//_ = levelEncoder.UnmarshalText([]byte(l.encoder.Level))

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        l.encoder.TimeKey,
		LevelKey:       levelKey,
		NameKey:        l.encoder.NameKey,
		CallerKey:      l.encoder.CallerKey,
		MessageKey:     l.encoder.MessageKey,
		StacktraceKey:  l.encoder.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    adminLevelEncoder,
		EncodeTime:     adminTimeEncoder,
		EncodeDuration: *durationEncoder,
		EncodeCaller:   *callerEncoder,
		EncodeName:     *nameEncoder,
	}

	return filterZapEncoder(l.encoder.Encoding, encoderConfig)
}
