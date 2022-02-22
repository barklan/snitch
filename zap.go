package snitch

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapSnitch struct {
	c    chan<- string
	conf *Config
	L    *zap.Logger
}

func OnZap(logger *zap.Logger, conf *Config) *ZapSnitch {
	c := make(chan string)
	go reporter(c, conf)
	return &ZapSnitch{
		c:    c,
		conf: conf,
		L:    logger,
	}
}

func (s *ZapSnitch) Info(msg string, fields ...zapcore.Field) {
	if s.conf.Level >= InfoLevel {
		s.c <- msg
	}
	s.L.Info(msg, fields...)
}

func (s *ZapSnitch) Warn(msg string, fields ...zapcore.Field) {
	if s.conf.Level >= WarningLevel {
		s.c <- msg
	}
	s.L.Warn(msg, fields...)
}

func (s *ZapSnitch) Error(msg string, fields ...zapcore.Field) {
	if s.conf.Level >= ErrorLevel {
		s.c <- msg
	}
	s.L.Error(msg, fields...)
}

func (s *ZapSnitch) Panic(msg string, fields ...zapcore.Field) {
	if s.conf.Level >= CriticalLevel {
		s.c <- msg
	}
	s.L.Panic(msg, fields...)
}

func (s *ZapSnitch) Fatal(msg string, fields ...zapcore.Field) {
	if s.conf.Level >= CriticalLevel {
		s.c <- msg
	}
	s.L.Fatal(msg, fields...)
}