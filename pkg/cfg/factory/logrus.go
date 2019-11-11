package factory

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/pkg/ext"
)

var Log = Option{
	Name:      "Log",
	OnCreate:  newLog,
	OnDestroy: flushLog,
}

func newLog(source Scanner) (interface{}, error) {
	var lc struct {
		Level    string   `json:"level"`
		Endpoint string   `json:"endpoint"`
		Key      string   `json:"key"`
		Secret   string   `json:"secret"`
		Project  string   `json:"project"`
		Name     string   `json:"name"`
		Topic    string   `json:"topic"`
		Extra    []string `json:"extra"`
		Default  bool     `json:"default"`
	}
	if err := source.Scan(&lc); err != nil {
		return nil, err
	}

	var (
		logger *logrus.Logger
		lsHook *ext.LogStoreHook
	)
	if lc.Default {
		logger = logrus.StandardLogger()
	} else {
		logger = logrus.New()
	}

	// Setup formatter
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	// Setup level
	level, err := logrus.ParseLevel(lc.Level)
	if err != nil {
		logrus.Warn("Invalid log level: ", lc.Level)
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if IsDefaultEnv() {
		return logger, nil
	}

	// Register hook
	var (
		extra   = sliceToMap(lc.Extra, "=")
		host, _ = os.Hostname()
	)
	lsHook, err = ext.NewLogStoreHook(lc.Endpoint, lc.Key, lc.Secret, lc.Project, lc.Name, lc.Topic, host, extra)
	if err != nil {
		return nil, err
	}
	logger.AddHook(lsHook)

	logrus.Debugf(
		"Setup logrus, project: %v, store: %v, topic: %v, host: %v, extra: %v",
		lc.Project, lc.Name, lc.Topic, host, extra)

	return logger, nil
}

func flushLog(v interface{}) {
	for _, hooks := range v.(*logrus.Logger).Hooks {
		for _, h := range hooks {
			if h, ok := h.(*ext.LogStoreHook); ok {
				_ = h.Flush(true)
			}
		}
	}
}
