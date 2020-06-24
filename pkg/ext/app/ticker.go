package app

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/pkg/ext/shutdown"
)

func RunTicker(name string, interval time.Duration, handler func()) {
	ticker := time.NewTicker(interval)
	ctx, cancel := context.WithCancel(context.Background())

	shutdown.AddHook(cancel)

	go func() {
		defer logrus.WithField("name", "Ticker ("+name+")").Debug("Stopped")
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				handler()
			}
		}
	}()
}
