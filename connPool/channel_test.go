package connPool

import (
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	InitialCap = 5
	MaximumCap = 30
	url        = "ws://39.108.191.63:40044/echo"
	orgin      = "http://localhost"
	factory    = newWsFactory(url, "", orgin)
)

type wsConnector struct {
	*websocket.Conn
}

func newWsConnector(conn *websocket.Conn) *wsConnector {
	comline := new(wsConnector)
	comline.Conn = conn
	return comline
}

func (this *wsConnector) SendRecv(buff []byte, timeout int) ([]byte, error) {
	var err error
	sendLen, err := this.Conn.Write(buff)
	fmt.Println(sendLen)
	if err != nil {
		return nil, err
	}

	err = this.Conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout)))
	if err != nil {
		return nil, err
	}

	recvBuff := make([]byte, 0)
	for {
		tmpBuff := make([]byte, 16)
		readLen, err := this.Conn.Read(tmpBuff)
		if err != nil {
			return nil, err
		}
		if readLen == 0 {
			break
		}
		if readLen < 16 {
			recvBuff = append(recvBuff, tmpBuff[0:readLen]...)
			break
		} else {
			recvBuff = append(recvBuff, tmpBuff...)
		}
	}

	if len(recvBuff) == 0 {
		return nil, errors.New("read websocket return empty buffer")
	}
	return recvBuff, nil
}

func (this *wsConnector) Close() error {
	return this.Conn.Close()
}

type wsFactory struct {
	url      string
	protocol string
	origin   string
}

func newWsFactory(url, protocol, origin string) *wsFactory {
	factory := new(wsFactory)
	factory.url = url
	factory.protocol = protocol
	factory.origin = origin
	return factory
}

func (this *wsFactory) Create() (Conn, error) {
	conn, err := websocket.Dial(this.url, this.protocol, this.origin)
	if err != nil {
		return nil, err
	}
	return newWsConnector(conn), nil

}

func TestNew(t *testing.T) {
	_, err := newChannelPool()
	if err != nil {
		t.Errorf("New error: %s", err.Error())
	}
}

//func TestPool_Get_Impl(t *testing.T) {
//	p, _ := newChannelPool()
//	defer p.Close()
//
//	conn, err := p.Get()
//	if err != nil {
//		t.Errorf("Get error: %s", err)
//	}
//
//	_, ok := conn.(*PoolConn)
//	if !ok {
//		t.Errorf("Conn is not of type poolConn")
//	}
//}

func TestPool_Get(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	_, err := p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}

	// after one get, current capacity should be lowered by one.
	if p.Len() != (InitialCap - 1) {
		t.Errorf("Get error. Expecting %d, got %d",
			(InitialCap - 1), p.Len())
	}

	// get them all
	var wg sync.WaitGroup
	for i := 0; i < (InitialCap - 1); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := p.Get()
			if err != nil {
				t.Errorf("Get error: %s", err)
			}
		}()
	}
	wg.Wait()

	if p.Len() != 0 {
		t.Errorf("Get error. Expecting %d, got %d",
			(InitialCap - 1), p.Len())
	}

	_, err = p.Get()
	if err != nil {
		t.Errorf("Get error: %s", err)
	}
}

func TestPool_Put(t *testing.T) {
	p, err := NewChannelPool(InitialCap, MaximumCap, factory)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	//startAt := time.Now()
	// get/Create from the pool
	conns := make([]*PoolConn, MaximumCap)
	for i := 0; i < MaximumCap; i++ {
		conn, _ := p.Get()
		conns[i] = conn
	}
	//fmt.Println("get all conn ", time.Since(startAt).Milliseconds())

	// now put them all back
	for _, conn := range conns {
		conn.BackClose()
	}

	if p.Len() != MaximumCap {
		t.Errorf("Put error len. Expecting %d, got %d",
			1, p.Len())
	}

	conn, _ := p.Get()
	p.Close() // close pool

	conn.Close() // try to put into a full pool
	if p.Len() != 0 {
		t.Errorf("Put error. Closed pool shouldn't allow to put connections.")
	}
}

func TestPool_PutUnusableConn(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	// ensure pool is not empty
	conn, _ := p.Get()
	conn.BackClose()

	poolSize := p.Len()
	conn, _ = p.Get()
	conn.BackClose()
	if p.Len() != poolSize {
		t.Errorf("Pool size is expected to be equal to initial size")
	}

	conn, _ = p.Get()
	//if pc, ok := conn.(*PoolConn); !ok {
	//	t.Errorf("impossible")
	//} else {
	//	pc.MarkUnusable()
	//}
	conn.MarkUnusable()
	conn.Close()
	if p.Len() != poolSize-1 {
		t.Errorf("Pool size is expected to be initial_size - 1,len[%d], size-1[%d]", p.Len(), poolSize-1)
	}
}

func TestPool_UsedCapacity(t *testing.T) {
	p, _ := newChannelPool()
	defer p.Close()

	if p.Len() != InitialCap {
		t.Errorf("InitialCap error. Expecting %d, got %d",
			InitialCap, p.Len())
	}
}

func TestPool_Close(t *testing.T) {
	p, _ := newChannelPool()

	// now close it and test all cases we are expecting.
	p.Close()

	c := p.(*channelPool)

	if c.conns != nil {
		t.Errorf("Close error, conns channel should be nil")
	}

	if c.factory != nil {
		t.Errorf("Close error, factory should be nil")
	}

	_, err := p.Get()
	if err == nil {
		t.Errorf("Close error, get conn should return an error")
	}

	if p.Len() != 0 {
		t.Errorf("Close error used capacity. Expecting 0, got %d", p.Len())
	}
}

func TestPoolConcurrent(t *testing.T) {
	p, _ := newChannelPool()
	pipe := make(chan *PoolConn, 0)

	go func() {
		p.Close()
	}()

	for i := 0; i < MaximumCap; i++ {
		go func() {
			conn, _ := p.Get()

			pipe <- conn
		}()

		go func() {
			conn := <-pipe
			if conn == nil {
				return
			}
			conn.BackClose()
		}()
	}
}

//func TestPoolWriteRead(t *testing.T) {
//	p, _ := NewChannelPool(0, 30, factory)
//
//	conn, _ := p.Get()
//
//	msg := "hello"
//	_, err := conn.Write([]byte(msg))
//	if err != nil {
//		t.Error(err)
//	}
//}

func TestPoolConcurrent2(t *testing.T) {
	p, _ := NewChannelPool(0, 30, factory)

	var wg sync.WaitGroup

	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				conn, _ := p.Get()
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				conn.Close()
				wg.Done()
			}(i)
		}
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			conn, _ := p.Get()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			conn.Close()
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func TestPoolConcurrent3(t *testing.T) {
	p, _ := NewChannelPool(0, 1, factory)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		p.Close()
		wg.Done()
	}()

	if conn, err := p.Get(); err == nil {
		conn.Close()
	}

	wg.Wait()
}

func newChannelPool() (Pool, error) {
	return NewChannelPool(InitialCap, MaximumCap, factory)
}

//
//func init() {
//	// used for factory function
//	go simpleWsServer()
//	time.Sleep(time.Millisecond * 300) // wait until tcp server has been settled
//
//	rand.Seed(time.Now().UTC().UnixNano())
//}
//
//func simpleWsServer() {
//
//	var err error
//	mux := http.NewServeMux()
//	mux.Handle("/gws", &wsHandler{})
//
//	err = http.ListenAndServe(":64005", mux)
//
//	if err != nil {
//		fmt.Println("g ws server start error", err)
//		return
//	}
//
//	return
//}
//
//type wsHandler struct {
//	MsgChannel chan string
//	locker     sync.Mutex
//	onceInit   sync.Once
//}
//
//var wsUpgrader = gws.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
//
////websocket的路径监听处理
//func (this *wsHandler) ServeHTTP(rew http.ResponseWriter, req *http.Request) {
//	conn, err := wsUpgrader.Upgrade(rew, req, nil)
//	if err != nil {
//		return
//	}
//
//	conn.SetPingHandler(func(appData string) error {
//		fmt.Println("ping handler")
//		return nil
//	})
//	conn.SetPongHandler(func(appData string) error {
//		fmt.Println("pong handler")
//		return nil
//	})
//	netAddr := conn.RemoteAddr()
//	netInfo := netAddr.String() //fmt.Sprintf("[%s]%s",netAddr.Network(),netAddr.String())
//	fmt.Println("ws conn enter:" + netInfo)
//	for {
//		msgType, msgBuff, err := conn.ReadMessage()
//		if err != nil {
//			fmt.Println("ws conn force close:" + netInfo)
//			conn.Close()
//			return
//		}
//
//		if 0 > msgType || gws.CloseMessage == msgType {
//			fmt.Println("ws conn grace close:" + netInfo)
//			conn.Close()
//			return
//		} else if gws.PingMessage == msgType {
//			conn.WriteMessage(gws.PongMessage, msgBuff)
//		} else if gws.TextMessage == msgType {
//			fmt.Printf("ws conn msg[%s]:%s\r\n", netInfo, string(msgBuff))
//			conn.WriteMessage(msgType, msgBuff)
//		} else if gws.BinaryMessage == msgType {
//			fmt.Printf("ws conn msg[%s]:%\r\n", netInfo, msgBuff)
//			conn.WriteMessage(msgType, msgBuff)
//		} else {
//		}
//	}
//
//	conn.Close()
//}

func TestPool_closeInvalidConn(t *testing.T) {
	p, err := NewChannelPool(MaximumCap, MaximumCap, factory)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	//startAt := time.Now()
	// get/Create from the pool
	conns := make([]*PoolConn, MaximumCap)
	for i := 0; i < MaximumCap; i++ {
		conn, _ := p.Get()
		conns[i] = conn
	}
	//fmt.Println("get all conn ", time.Since(startAt).Milliseconds())

	// now put them all back
	for _, conn := range conns {
		_, err := conn.SendRecv([]byte("this is a test hello world"), 30000)
		if err != nil {
			printErr(err)
		}
		conn.BackClose()
	}

	if p.Len() != MaximumCap {
		t.Errorf("Put error len. Expecting %d, got %d",
			1, p.Len())
	}

	conn, _ := p.Get()
	p.Close() // close pool

	conn.Close() // try to put into a full pool
	if p.Len() != 0 {
		t.Errorf("Put error. Closed pool shouldn't allow to put connections.")
	}
}

func printErr(err error) {
	fmt.Println(err)
	if err == io.EOF {
		fmt.Println("io.EOF")
		return
	}
	netErr, ok := err.(net.Error)
	if ok {
		if netErr.Timeout() {
			fmt.Println("netErr timeout")
		} else if netErr.Temporary() {
			fmt.Println("netErr Temporary")
		}
	}

	opErr, ok := netErr.(*net.OpError)
	if ok {
		switch t := opErr.Err.(type) {
		case *net.DNSError:
			fmt.Printf("net.DNSError:%+v", t)
		case *os.SyscallError:
			fmt.Printf("os.SyscallError:%+v", t)
			if errno, ok := t.Err.(syscall.Errno); ok {
				switch errno {
				case syscall.ECONNREFUSED:
					fmt.Println("connect refused")
				case syscall.ETIMEDOUT:
					fmt.Println("syscall etimeout")
				}
			}
		default:
			fmt.Println("OpError unknow")
		}
	}

	fmt.Println("error unknow")
}
