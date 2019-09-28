package factory

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"skeleton/pkg/ext"
)

var components = make(map[string]interface{})

func Register(name string, option Option, source Scanner, lazy bool) ProviderMethod {
	var (
		once     = &sync.Once{}
		fullName = fmt.Sprintf("%s (%s)", option.Name, name)
		logger   = logrus.WithField("name", fullName)
	)

	// Each provider should init only once
	init := func() {
		if option.OnCreate == nil {
			return
		}

		once.Do(func() {
			logger.Debug("Loading...")
			value, err := option.OnCreate(source)
			if err != nil {
				components[fullName] = xerrors.Errorf("while load %s: %w", fullName, err)
				return
			}
			logger.Debug("Loaded")

			if option.OnCreated != nil {
				option.OnCreated(fullName, value)
			}

			if option.OnDestroy != nil {
				// Register shutdown hook
				ext.OnShutdown(func() {
					logger.Debug("Start shutdown...")
					option.OnDestroy(value)
					logger.Debug("Finish Shutdown")
				})
			}

			components[fullName] = value
		})

		if v, ok := components[fullName]; !ok {
			logger.Fatalf("Component not found")
		} else if err, ok := v.(error); ok {
			logger.WithError(err).Fatal("Fail init")
		}
	}

	// Eager load
	if !lazy {
		init()
	}

	return func(target interface{}) {
		init()

		sv := reflect.ValueOf(components[fullName])
		tv := reflect.ValueOf(target).Elem()
		tv.Set(sv)
	}
}
