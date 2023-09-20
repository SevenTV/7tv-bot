package ratelimit

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisClient returns a redis client configured with the given options, ready for use with the RateLimiter
func RedisClient(opt RedisOptions) (*redis.Client, error) {
	var client *redis.Client

	if len(opt.Addresses) == 0 {
		return nil, fmt.Errorf("you must provide at least one redis address")
	}

	if opt.Sentinel {
		client = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       opt.MasterName,
			SentinelAddrs:    opt.Addresses,
			SentinelUsername: opt.Username,
			SentinelPassword: opt.Password,
			Username:         opt.Username,
			Password:         opt.Password,
			DB:               opt.Database,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     opt.Addresses[0],
			Username: opt.Username,
			Password: opt.Password,
			DB:       opt.Database,
		})
	}

	return client, nil
}

// RedisOptions contain config variables for a redis client
type RedisOptions struct {
	MasterName string
	Username   string
	Password   string
	Database   int

	Addresses []string
	Sentinel  bool

	EnableSync bool
}
