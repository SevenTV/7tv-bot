package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/gookit/goutil/mathutil"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	joinKey = "twitch-irc-join-ratelimit"
	authKey = "twitch-irc-auth-ratelimit"
)

type RateLimiter struct {
	redisClient          *redis.Client
	mx                   *sync.Mutex
	joinLimit, authLimit int64
	reset                time.Duration
}

// New returns a new RateLimiter configured with the passed parameters
func New(rc *redis.Client, joinLimit, authLimit int64, reset time.Duration) *RateLimiter {
	return &RateLimiter{
		redisClient: rc,
		mx:          &sync.Mutex{},
		joinLimit:   joinLimit,
		authLimit:   authLimit,
		reset:       reset,
	}
}

// KeepAlive sends repeated pings to Redis, gives an error when ping fails, so you know the connection died
func (r *RateLimiter) KeepAlive(ctx context.Context) error {
	for range time.NewTicker(10 * time.Second).C {
		err := r.redisClient.Ping(ctx).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RateLimiter) getRate(key string, ctx context.Context) (counter int64, err error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	counter, err = r.redisClient.Get(ctx, key).Int64()
	if err != nil {
		if err != redis.Nil {
			return
		}
		err = r.redisClient.Set(ctx, key, 0, r.reset).Err()
		if err != nil {
			return
		}
	}
	counter, err = r.redisClient.Incr(ctx, key).Result()
	return
}

// WaitToJoin is a blocking function that returns when we have capacity in the rate limit to Join a channel
func (r *RateLimiter) WaitToJoin(ctx context.Context) error {
	count, err := r.getRate(joinKey, ctx)
	zap.S().Debugf("join ratelimit: %v", count)
	if err != nil {
		return err
	}
	if count > r.joinLimit {
		ttl, _ := r.redisClient.TTL(ctx, joinKey).Result()
		<-time.After(addJitter(ttl))
		err = r.WaitToJoin(ctx)
	}
	return err
}

// WaitToAuth is a blocking function that returns when we have capacity in the rate limit to create a new IRC connection
func (r *RateLimiter) WaitToAuth(ctx context.Context) error {
	count, err := r.getRate(authKey, ctx)
	zap.S().Debugf("auth ratelimit: %v", count)
	if err != nil {
		return err
	}
	if count > r.joinLimit {
		ttl, _ := r.redisClient.TTL(ctx, authKey).Result()
		<-time.After(addJitter(ttl))
		err = r.WaitToAuth(ctx)
	}
	return err
}

// we want to add a little jitter before retrying the join/auth, so we don't send thousands of requests to redis at the same moment
func addJitter(duration time.Duration) time.Duration {
	return duration + time.Duration(mathutil.RandInt(50, 2000))*time.Millisecond
}
