package snitch

import (
	"testing"

	mock_snitch "github.com/barklan/snitch/mock"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

func mockPartialBack(t *testing.T) (*gomock.Controller, chan string, *Config, *mock_snitch.Mockbot) {
	t.Helper()

	ctrl := gomock.NewController(t)

	m := mock_snitch.NewMockbot(ctrl)
	conf := &Config{
		TGToken:  "abc",
		TGChatID: 342495235534,
		Level:    ErrorLevel,
	}

	c := make(chan string, 10)
	return ctrl, c, conf, m
}

func TestZap(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("failed to init zap logger")
	}

	ctrl, c, conf, m := mockPartialBack(t)
	defer ctrl.Finish()
	snitch := &ZapSnitch{
		c:    c,
		conf: conf,
		L:    logger,
	}

	msg := "this is a log message"

	// Asserts that the first and only call to Bar() is passed 99.
	// Anything else will fail.
	m.
		EXPECT().
		ChatByID(gomock.Eq(conf.TGChatID))

	m.
		EXPECT().Handle(gomock.Any(), gomock.Any())

	m.
		EXPECT().
		Send(gomock.Any(), gomock.Eq(msg)).MaxTimes(1)

	back, err := newBackend(conf, m, c)
	if err != nil {
		t.Errorf("failed to init backend")
	}

	go back.start()

	snitch.Error(msg, zap.String("test", "test"))
	snitch.Error(msg, zap.String("test", "test"))
}
