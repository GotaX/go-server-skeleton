package factory

import (
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
		Async    bool     `json:"async"`
	}
	if err := source.Scan(&lc); err != nil {
		return nil, err
	}

	var (
		logger *logrus.Logger
		lsHook logrus.Hook
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
	c := ext.LogStoreConfig{
		Endpoint:     lc.Endpoint,
		AccessKey:    lc.Key,
		AccessSecret: lc.Secret,
		Project:      lc.Project,
		Store:        lc.Name,
		Topic:        lc.Topic,
		Source:       ext.HostName(),
		Extra:        sliceToMap(lc.Extra, "="),
	}

	c.Extra["version"] = ext.Version()

	if lc.Async {
		lsHook, err = ext.NewAsyncLogStoreHook(c)
	} else {
		lsHook, err = ext.NewLogStoreHook(c)
	}
	if err != nil {
		return nil, err
	}
	logger.AddHook(lsHook)

	logrus.Debugf(
		"Setup logrus, project: %v, store: %v, topic: %v, host: %v, extra: %v",
		c.Project, c.Store, c.Topic, c.Source, c.Extra)

	return logger, nil
}

func flushLog(v interface{}) {
	for _, hooks := range v.(*logrus.Logger).Hooks {
		for _, h := range hooks {
			switch h := h.(type) {
			case *ext.LogStoreHook:
				_ = h.Flush(true)
			case *ext.AsyncLogStoreHook:
				_ = h.Close()
			}
		}
	}
}
