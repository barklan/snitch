package snitch

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapSnitch struct {
	c    chan<- string
	Conf *Config
	L    *zap.Logger
}

func OnZap(logger *zap.Logger, conf *Config) (*ZapSnitch, error) {
	c := make(chan string, 10)
	b, err := newBot(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to init tg bot: %w", err)
	}
	back, err := newBackend(conf, b, c)
	if err != nil {
		return nil, fmt.Errorf("failed to start backend: %w", err)
	}
	go back.start()
	return &ZapSnitch{
		c:    c,
		Conf: conf,
		L:    logger,
	}, nil
}

func (s *ZapSnitch) Debug(msg string, fields ...zapcore.Field) {
	s.L.Debug(msg, fields...)
}

func (s *ZapSnitch) Info(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= InfoLevel {
		s.c <- msg
	}
	s.L.Info(msg, fields...)
}

func (s *ZapSnitch) Warn(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= WarnLevel {
		s.c <- msg
	}
	s.L.Warn(msg, fields...)
}

func (s *ZapSnitch) Error(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= ErrorLevel {
		s.c <- msg
	}
	s.L.Error(msg, fields...)
}

func (s *ZapSnitch) Panic(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= CritLevel {
		s.c <- msg
	}
	s.L.Panic(msg, fields...)
}

func (s *ZapSnitch) Fatal(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= CritLevel {
		s.c <- msg
	}
	s.L.Fatal(msg, fields...)
}
