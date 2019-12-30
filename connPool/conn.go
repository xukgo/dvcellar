package connPool

import (
	"sync"
)

// PoolConn是一个对conn到包装类，额外添加了锁和关联的pool
type PoolConn struct {
	Conn
	mu       sync.RWMutex
	c        *channelPool
	unusable bool
}

func (p *PoolConn) GetConn() Conn {
	return p.Conn
}
//如果有效到连接则放回pool，无效的则关闭掉
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

// 标记连接是否还有效否
func (p *PoolConn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

func (c *channelPool) wrapConn(conn Conn) *PoolConn {
	p := &PoolConn{c: c}
	p.Conn = conn
	return p
}
