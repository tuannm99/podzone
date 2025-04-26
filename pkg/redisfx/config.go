package redisfx

import (
	"net/url"
	"strconv"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit"
)

type RedisSetting struct {
	Addr          string
	Password      string
	DB            int
	Timeout       time.Duration
	RetryAttempts int
}

func NewRedisSetting() RedisSetting {
	addr := toolkit.FallbackEnv("REDIS_ADDR", "redis://localhost:6379/0")

	redisUrl, _ := url.Parse(addr)
	pass, _ := redisUrl.User.Password()
	db := 0
	if redisUrl.Path != "" {
		// remove leading `/` from path and parse as int
		if parsed, err := strconv.Atoi(redisUrl.Path[1:]); err == nil {
			db = parsed
		}
	}

	return RedisSetting{
		Addr:          redisUrl.Host,
		Password:      pass,
		DB:            db,
		Timeout:       3 * time.Second,
		RetryAttempts: 10,
	}
}
