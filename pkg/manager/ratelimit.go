package manager

import "context"

// RateLimiter is an interface to be used to implement a rate limiter for the IRC manager
type RateLimiter interface {
	// WaitToJoin blocks until capacity is available in the rate limit
	WaitToJoin(ctx context.Context) error
	// WaitToAuth blocks until capacity is available in the rate limit
	WaitToAuth(ctx context.Context) error
}

// NoLimit implements the RateLimiter interface, to be used as default value in the IRC manager
type NoLimit struct{}

// WaitToJoin is a no-op
func (_ NoLimit) WaitToJoin(_ context.Context) error {
	return nil
}

// WaitToAuth is a no-op
func (_ NoLimit) WaitToAuth(_ context.Context) error {
	return nil
}
