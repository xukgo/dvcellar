package connPool

import (
	"errors"
	"fmt"
	"sync"
)

type channelPool struct {
	//storage conn connections
	mu    sync.RWMutex
	conns chan Conn

	// Conn generator factory define
	factory ConnFactory
}

func NewChannelPool(initialCap, maxCap int, factory ConnFactory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	p := &channelPool{
		conns:   make(chan Conn, maxCap),
		factory: factory,
	}

	// 初始化链接，如果失败，关闭pool 返回错误
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := factory.Create()
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err.Error())
		}
		p.conns <- conn
	}

	return p, nil
}

func (c *channelPool) getConnsAndFactory() (chan Conn, ConnFactory) {
	c.mu.RLock()
	conns := c.conns
	factory := c.factory
	c.mu.RUnlock()
	return conns, factory
}

func (c *channelPool) Get() (*PoolConn, error) {
	conns, factory := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}

	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		return c.wrapConn(conn), nil
	default:
		conn, err := factory.Create()
		if err != nil {
			return nil, err
		}

		return c.wrapConn(conn), nil
	}
}

func (c *channelPool) put(conn Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conns == nil {
		// 如果pool关闭了，关闭传过来的连接
		return conn.Close()
	}

	//把连接放回到pool到时候要分情况，如果pool满了，则会阻塞住执行default，把连接关闭关闭
	select {
	case c.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	conns, _ := c.getConnsAndFactory()
	return len(conns)
}
