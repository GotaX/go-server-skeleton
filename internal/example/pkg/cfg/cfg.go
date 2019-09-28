package cfg

import (
	"strings"

	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/env"
	"github.com/sirupsen/logrus"

	"skeleton/pkg/cfg/factory"
	"skeleton/pkg/ext"
	"skeleton/pkg/ext/config/spring"
)

const (
	sApp               = "app"
	kAppName, dAppName = "name", "demo"
	kProfile, dProfile = factory.EnvProfile, factory.DefaultProfile
)

var (
	LogGrpc factory.ProviderMethod
)

func init() {
	name, profile := loadConfig()

	_ = register("log", factory.Log, false)
	LogGrpc = register("logGrpc", factory.Log, false)
	_ = register("trace", factory.Tracing, false)

	logrus.Infof("Init over, profile: %s-%s\n", name, profile)
}

func register(name string, create factory.Option, lazy bool) factory.ProviderMethod {
	source := config.Get(strings.Split(name, ".")...)
	return factory.Register(name, create, source, lazy)
}

func loadConfig() (name, profile string) {
	if err := config.Load(env.NewSource()); err != nil {
		logrus.Fatal(err)
	}
	ext.OnShutdown(func() {
		entry := logrus.WithField("name", "ConfigWatcher")
		if err := config.DefaultConfig.Close(); err != nil {
			entry.WithError(err).Warn("Fail to stop")
		}
		entry.Debug("Stopped")
	})

	name = config.Get(sApp, kAppName).String(dAppName)
	profile = config.Get(sApp, kProfile).String(dProfile)

	if err := config.Load(spring.NewSource(
		spring.WithName(name),
		spring.WithProfile(profile),
	)); err != nil {
		logrus.WithError(err).Fatal("Exit with error")
	}
	return
}
