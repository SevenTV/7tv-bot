package pubsub

import "sync"

var (
	connections = make(map[*Connection]bool)
	mx          *sync.Mutex
)

func Init() {
	mx = &sync.Mutex{}
}

func Publish(data []byte) {
	mx.Lock()
	defer mx.Unlock()
	for conn := range connections {
		conn.Write(data)
	}
}
