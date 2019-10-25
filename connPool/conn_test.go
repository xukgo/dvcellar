package connPool

import (
	"fmt"
	"testing"
)

func TestConn_Impl(t *testing.T) {
	// Create a factory() to be used with channel based pool
	factory := newWsFactory(url, "", orgin)

	// Create a new channel based pool with an initial capacity of 5 and maximum
	// capacity of 30. The factory will Create 5 initial connections and put it
	// into the pool.
	p, err := NewChannelPool(5, 30, factory)
	if err != nil {
		t.Fail()
	}

	// now you can get a connection from the pool, if there is no connection
	// available it will Create a new one via the factory function.
	conn, err := p.Get()

	// do something with conn and put it back to the pool by closing the connection
	// (this doesn't close the underlying connection instead it's putting it back
	// to the pool).
	conn.Close()

	// close the underlying connection instead of returning it to pool
	// it is useful when acceptor has already closed connection and conn.Write() returns error
	//if pc, ok := conn.(*PoolConn); ok {
	//	pc.MarkUnusable()
	//	pc.Close()
	//}
	conn.MarkUnusable()
	conn.BackClose()

	// close pool any time you want, this closes all the connections inside a pool
	p.Close()

	// currently available connections in the pool
	current := p.Len()
	fmt.Println(current)
}
