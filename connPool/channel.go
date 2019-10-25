package connPool

import (
	"errors"
	"fmt"
	"sync"
)

type channelPool struct {
	// storage for our Conn connections
	mu    sync.RWMutex
	conns chan Conn

	// Conn generator
	factory ConnFactory
}

func NewChannelPool(initialCap, maxCap int, factory ConnFactory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		conns:   make(chan Conn, maxCap),
		factory: factory,
	}

	// Create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := factory.Create()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err.Error())
		}
		c.conns <- conn
	}

	return c, nil
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
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		return nil
	default:
		// pool is full, close passed connection
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
