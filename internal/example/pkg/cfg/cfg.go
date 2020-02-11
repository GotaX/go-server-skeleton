package cfg

import (
	"strings"

	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/env"
	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
	logrus2 "github.com/GotaX/go-server-skeleton/pkg/cfg/logrus"
	"github.com/GotaX/go-server-skeleton/pkg/cfg/tracing"
	"github.com/GotaX/go-server-skeleton/pkg/ext/config/spring"
	"github.com/GotaX/go-server-skeleton/pkg/ext/shutdown"
)

const (
	sApp               = "app"
	kAppName, dAppName = "name", "demo"
	kProfile, dProfile = "profile", "default"
)

var (
	LogGrpc cfg.ProviderMethod
)

func init() {
	name, profile := loadConfig()

	_ = register("log", logrus2.Option, false)
	LogGrpc = register("logGrpc", logrus2.Option, false)
	_ = register("trace", tracing.Option, false)

	logrus.Infof("Init over, profile: %s-%s\n", name, profile)
}

func register(name string, create cfg.Option, lazy bool) cfg.ProviderMethod {
	source := config.Get(strings.Split(name, ".")...)
	return cfg.Register(name, create, source, lazy)
}

func loadConfig() (name, profile string) {
	if err := config.Load(env.NewSource()); err != nil {
		logrus.Fatal(err)
	}
	shutdown.AddHook(func() {
		entry := logrus.WithField("name", "ConfigWatcher")
		if err := config.DefaultConfig.Close(); err != nil {
			entry.WithError(err).Warn("Fail to stop")
		}
		entry.Debug("Stopped")
	})

	name = config.Get(sApp, kAppName).String(dAppName)
	profile = config.Get(sApp, kProfile).String(dProfile)

	if err := config.Load(
		spring.NewSource(
			spring.WithName(name),
			spring.WithProfile(profile),
		),
		env.NewSource(),
	); err != nil {
		logrus.WithError(err).Fatal("Exit with error")
	}
	return
}
