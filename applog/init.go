package applog

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Init logging entry point
func Init(isVerbose bool) {
	f := &MultiLineFormatter{
		TextFormatter: log.TextFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000000",
			FullTimestamp:   true,
		},
	}
	log.SetFormatter(f)

	logLevel := log.InfoLevel
	if isVerbose {
		logLevel = log.DebugLevel
	}

	log.SetLevel(logLevel)

	log.Debugf("will set log level to %v", logLevel)
}

type MultiLineFormatter struct {
	log.TextFormatter
}

func (f *MultiLineFormatter) Format(entry *log.Entry) ([]byte, error) {
	multiline, ok := entry.Data["multiline"]
	if ok {
		delete(entry.Data, "multiline")
	}
	res, err := f.TextFormatter.Format(entry)
	if multiline, ok := multiline.(string); ok && multiline != "" {
		res = append(res, []byte(multiline)...)
	}
	return res, err
}

type BufferedLogs struct {
	Messages []string
}

func (bl *BufferedLogs) Fire(entry *log.Entry) error {
	bl.Messages = append(bl.Messages, entry.Message)
	if len(entry.Data) > 0 {
		for _, dataItem := range entry.Data {
			bl.Messages = append(bl.Messages, fmt.Sprint(dataItem))
		}
	}

	return nil
}

func (bl *BufferedLogs) Levels() []log.Level {
	return log.AllLevels
}
