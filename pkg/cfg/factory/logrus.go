package factory

import (
	"context"
	"time"

	slsh "github.com/GotaX/logrus-aliyun-log-hook"
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

	extra := sliceToMap(lc.Extra, "=")
	extra["version"] = ext.Version()

	hook, err := slsh.New(slsh.Config{
		Endpoint:     lc.Endpoint,
		AccessKey:    lc.Key,
		AccessSecret: lc.Secret,
		Project:      lc.Project,
		Store:        lc.Name,
		Topic:        lc.Topic,
		Extra:        extra,
	})
	if err != nil {
		return nil, err
	}
	logger.AddHook(hook)
	return logger, nil
}

func flushLog(v interface{}) {
	for _, hooks := range v.(*logrus.Logger).Hooks {
		for _, h := range hooks {
			switch h := h.(type) {
			case *slsh.Hook:
				ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
				_ = h.CloseContext(ctx)
			}
		}
	}
}
