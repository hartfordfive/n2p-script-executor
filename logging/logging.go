package logging

import (
	log "github.com/sirupsen/logrus"
)

type UTCLoggingFormatter struct {
	log.Formatter
}

var LogFormatJson = UTCLoggingFormatter{&log.JSONFormatter{
	TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	FieldMap: log.FieldMap{
		log.FieldKeyTime:  "@timestamp",
		log.FieldKeyLevel: "level",
		log.FieldKeyMsg:   "message",
		log.FieldKeyFunc:  "caller",
	},
}}

var LogFormatPlain = log.TextFormatter{
	TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
	DisableColors:   true,
	FullTimestamp:   true,
	FieldMap: log.FieldMap{
		log.FieldKeyTime:  "timestamp",
		log.FieldKeyLevel: "level",
		log.FieldKeyMsg:   "message",
		log.FieldKeyFunc:  "caller",
	},
}

func (u UTCLoggingFormatter) Format(e *log.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func SetLogLevel(lvl string) {
	if lvl == "error" {
		log.SetLevel(log.ErrorLevel)
	} else if lvl == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if lvl == "info" {
		log.SetLevel(log.InfoLevel)
	} else if lvl == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}
