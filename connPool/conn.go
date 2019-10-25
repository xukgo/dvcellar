package connPool

import (
	"sync"
)

// PoolConn is a wrapper around *websocket.Conn to modify the the behavior of
// *websocket.Conn's Close() method.
type PoolConn struct {
	Conn
	mu       sync.RWMutex
	c        *channelPool
	unusable bool
}

// Close() puts the given connects back to the pool instead of closing it.
func (p *PoolConn) BackClose() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.unusable {
		if p.Conn != nil {
			return p.Conn.Close()
		}
		return nil
	}
	return p.c.put(p.Conn)
}

// MarkUnusable() marks the connection not usable any more, to let the pool close it instead of returning it to pool.
func (p *PoolConn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

// newConn wraps a standard *websocket.Conn to a poolConn *websocket.Conn.
func (c *channelPool) wrapConn(conn Conn) *PoolConn {
	p := &PoolConn{c: c}
	p.Conn = conn
	return p
}

//func RealClose(){
//	if pc, ok := conn.(*PoolConn); !ok {
//		t.Errorf("impossible")
//	} else {
//		pc.MarkUnusable()
//	}
//	conn.Close()
//}
