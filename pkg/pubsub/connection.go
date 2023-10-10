package pubsub

type Connection struct {
	cb func([]byte)
}

func NewConnection(cb func([]byte)) *Connection {
	conn := &Connection{
		cb: cb,
	}
	connections[conn] = true
	return conn
}

func (c *Connection) Write(data []byte) {
	c.cb(data)
}

func (c *Connection) Close() {
	mx.Lock()
	defer mx.Unlock()
	c.cb = nil
	delete(connections, c)
}
