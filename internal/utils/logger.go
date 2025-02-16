package utils

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger() *logrus.Entry {
	lumberjack := &lumberjack.Logger{
		Filename:   "./alchemy.log",
		MaxSize:    1,
		MaxBackups: 0,
		MaxAge:     0,
		LocalTime:  true,
	}

	log := logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.ErrorLevel)
	log.SetOutput(lumberjack)

	return log.WithField("sessionId", lo.RandomString(16, lo.AlphanumericCharset))
}
