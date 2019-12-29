package unitTest

import (
	"github.com/xukgo/dvcellar/connPool"
	"testing"
)

func TestConn_Impl(t *testing.T) {
	factory := newWsFactory(url, "", orgin, 1000)

	p, err := connPool.NewChannelPool(5, 30, factory)
	if err != nil {
		t.Fail()
		return
	}

	conn, err := p.Get()

	conn.Close()
	conn.MarkUnusable()
	conn.BackClose()

	if p.Len() != 4 {
		t.Fail()
	}

	p.Close()

	current := p.Len()

	if current != 0 {
		t.Fail()
	}
}
