package spring

import (
	"context"

	"github.com/micro/go-micro/config/source"
)

func WithProfile(profile string) source.Option {
	return withValue(kConfProfile, profile)
}

func WithLabel(label string) source.Option {
	return withValue(kConfLabel, label)
}

func WithHost(host string) source.Option {
	return withValue(kConfUrl, host)
}

func WithName(name string) source.Option {
	return withValue(kConfName, name)
}

func withValue(key, value interface{}) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, key, value)
	}
}
