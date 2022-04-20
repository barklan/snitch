package snitch

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Zap struct {
	c    chan<- string
	Conf *Config
	bot  bot
	L    *zap.Logger
}

// Function OnZap constructs wrapper around zap.Logger.
func OnZap(logger *zap.Logger, conf *Config) (*Zap, error) {
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
	return &Zap{
		c:    c,
		Conf: conf,
		bot:  b,
		L:    logger,
	}, nil
}

// Debug directly calls logger.Debug().
func (s *Zap) Debug(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= DebugLevel && msg != "" {
		select {
		case s.c <- DebugPrefix + msg:
		default:
			s.L.Debug("channel buzy, message dropped")
		}
	}
	s.L.Debug(msg, fields...)
}

func (s *Zap) Info(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= InfoLevel && msg != "" {
		select {
		case s.c <- InfoPrefix + msg:
		default:
			s.L.Info("channel buzy, message dropped")
		}
	}
	s.L.Info(msg, fields...)
}

func (s *Zap) Warn(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= WarnLevel && msg != "" {
		select {
		case s.c <- WarnPrefix + msg:
		default:
			s.L.Warn("channel buzy, message dropped")
		}
	}
	s.L.Warn(msg, fields...)
}

func (s *Zap) Error(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= ErrorLevel && msg != "" {
		select {
		case s.c <- ErrorPrefix + msg:
		default:
			s.L.Error("channel buzy, message dropped")
		}
	}
	s.L.Error(msg, fields...)
}

// Panic snitches if level <= CritLevel and calls logger.Panic().
func (s *Zap) Panic(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= CritLevel {
		s.allHellBrokeLoose(msg)
	}
	s.L.Panic(msg, fields...)
}

// Fatal snitches if level <= CritLevel and calls logger.Fatal().
func (s *Zap) Fatal(msg string, fields ...zapcore.Field) {
	if s.Conf.Level <= CritLevel {
		s.allHellBrokeLoose(msg)
	}
	s.L.Fatal(msg, fields...)
}

func (s *Zap) allHellBrokeLoose(msg string) {
	chat, err := s.bot.ChatByID(s.Conf.TGChatID)
	if err != nil {
		if _, e := s.bot.Send(chat, CritPrefix+msg); e != nil {
			time.Sleep(50 * time.Millisecond)
			_, _ = s.bot.Send(chat, "CRITICAL!")
		}
	}
}
