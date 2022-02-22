package snitch

import (
	"testing"
	"time"

	mock_snitch "github.com/barklan/snitch/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func mockPartialBack(t *testing.T) (*gomock.Controller, chan string, *Config, *mock_snitch.Mockbot) {
	t.Helper()

	ctrl := gomock.NewController(t)

	m := mock_snitch.NewMockbot(ctrl)
	conf := &Config{
		TGToken:   "abc",
		TGChatID:  23423344,
		Level:     InfoLevel,
		Cooldown:  50 * time.Millisecond,
		CacheSize: 5,
	}

	c := make(chan string, 10)
	return ctrl, c, conf, m
}

func TestZap(t *testing.T) {
	t.Run("twice the same message should send only one", func(t *testing.T) {
		ctrl, c, conf, m := mockPartialBack(t)
		defer ctrl.Finish()

		logger, err := zap.NewDevelopment()
		if err != nil {
			t.Errorf("failed to init zap logger")
		}

		snitch := &Zap{
			c:    c,
			Conf: conf,
			bot:  m,
			L:    logger,
		}

		msg := "this is a log message"
		m.EXPECT().ChatByID(gomock.Eq(conf.TGChatID))
		m.EXPECT().Handle(gomock.Any(), gomock.Any())
		m.EXPECT().Start()
		m.EXPECT().Send(gomock.Any(), gomock.Eq(ErrorPrefix+msg)).AnyTimes()

		back, err := newBackend(conf, m, c)
		if err != nil {
			t.Errorf("failed to init backend")
		}

		go back.start()
		time.Sleep(50 * time.Millisecond)
		snitch.Error(msg, zap.String("test", "test"))
		snitch.Error(msg, zap.String("test", "test"))
	})
	t.Run("twice the same message after cooldown", func(t *testing.T) {
		ctrl, c, conf, m := mockPartialBack(t)
		defer ctrl.Finish()

		logger, err := zap.NewDevelopment()
		if err != nil {
			t.Errorf("failed to init zap logger")
		}

		snitch := &Zap{
			c:    c,
			Conf: conf,
			bot:  m,
			L:    logger,
		}

		msg := "this is a log message"
		m.EXPECT().ChatByID(gomock.Eq(conf.TGChatID))
		m.EXPECT().Handle(gomock.Any(), gomock.Any())
		m.EXPECT().Start()
		m.EXPECT().Send(gomock.Any(), gomock.Any()).Times(4)

		back, err := newBackend(conf, m, c)
		if err != nil {
			t.Errorf("failed to init backend")
		}

		go back.start()
		time.Sleep(50 * time.Millisecond)
		snitch.Error(msg, zap.String("test", "test"))
		time.Sleep(80 * time.Millisecond)
		snitch.Error(msg, zap.String("test", "test"))
		time.Sleep(80 * time.Millisecond)
		snitch.Warn("another message")
		time.Sleep(80 * time.Millisecond)
		snitch.Info("and info message")
		time.Sleep(80 * time.Millisecond)
		snitch.Debug("debug message should not be sent")
	})
	t.Run("send critical notification in sync", func(t *testing.T) {
		ctrl, c, conf, m := mockPartialBack(t)
		defer ctrl.Finish()

		logger, err := zap.NewDevelopment()
		if err != nil {
			t.Errorf("failed to init zap logger")
		}

		snitch := &Zap{
			c:    c,
			Conf: conf,
			bot:  m,
			L:    logger,
		}

		msg := "some critical error"
		m.EXPECT().ChatByID(gomock.Eq(conf.TGChatID))
		m.EXPECT().Handle(gomock.Any(), gomock.Any())
		m.EXPECT().Start()
		m.EXPECT().ChatByID(gomock.Eq(conf.TGChatID))

		back, err := newBackend(conf, m, c)
		if err != nil {
			t.Errorf("failed to init backend")
		}

		go back.start()
		time.Sleep(50 * time.Millisecond)
		assert.Panics(
			t,
			func() {
				snitch.Panic(msg, zap.String("test", "test"))
			})
	})
}
