package main

import (
	"testing"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpostgres"
	"github.com/tuannm99/podzone/pkg/pdredis"
)

func TestMain(t *testing.T) {
	t.Setenv("CONFIG_PATH", "config.yml")
	pdlog.Registry.Use("noop")
	pdredis.Registry.Use("noop")
	pdpostgres.Registry.Use("noop")

	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Log("main() still running, test will stop here")
	}
}
