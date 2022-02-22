package snitch

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapSnitch struct {
	c    chan<- string
	conf *Config
	L    *zap.Logger
}

func OnZap(logger *zap.Logger, conf *Config) (*ZapSnitch, error) {
	c := make(chan string, 10)
	// FIXME init bot here and return err if needed
	back, err := newBackend(conf, c)
	if err != nil {
		return nil, fmt.Errorf("failed to start backend: %w", err)
	}
	go back.Start()
	return &ZapSnitch{
		c:    c,
		conf: conf,
		L:    logger,
	}, nil
}

func (s *ZapSnitch) Debug(msg string, fields ...zapcore.Field) {
	s.L.Debug(msg, fields...)
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
