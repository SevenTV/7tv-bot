package util

import "sync"

// Closer is useful for signaling
type Closer struct {
	mx sync.Mutex
	o  *sync.Once
	// C is the channel that gets closed with Close
	C chan struct{}
}

func (c *Closer) Reset() {
	c.mx.Lock()
	c.o = &sync.Once{}
	c.C = make(chan struct{})
	c.mx.Unlock()
}

func (c *Closer) Close() {
	c.mx.Lock()
	c.o.Do(func() {
		close(c.C)
	})
	c.mx.Unlock()
}
