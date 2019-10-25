//这里需要ping pong心跳来保活侦测这个链接，否则并发抢占新建连接会出问题
package comLine

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type KeepAliveWebsocketClient struct {
	client  *websocket.Conn
	url     string
	locker  sync.Mutex
	isvalid bool
	exit    bool
}

func NewKeepAliveWebsocketClient(inUrl string) *KeepAliveWebsocketClient {
	return &KeepAliveWebsocketClient{
		url:     inUrl,
		isvalid: false,
		exit:    false,
	}
}

func (this *KeepAliveWebsocketClient) IsValid() bool {
	return this.isvalid
}

//开始保持连接，断线会自动重连，该函数会阻塞
func (this *KeepAliveWebsocketClient) StartKeepAlive(timeout int) {
	var err error
	var heartbeatInterval int64 = 2000
	tryCount := 2 //pingpong失败几次认定连接已经失效
	errCount := 0

	for {
		//连接
		if !this.isvalid {
			err = this.connect(timeout)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			time.Sleep(time.Second)
		}

		startTime := time.Now()
		//ping pong机制
		err = this.pingPongCheck(timeout)
		endTime := time.Now()
		duration := (endTime.UnixNano() - startTime.UnixNano()) / 1000000
		if duration < heartbeatInterval {
			time.Sleep(time.Millisecond * time.Duration(heartbeatInterval-duration))
		}

		if err != nil {
			errCount++
			if errCount >= tryCount {
				this.isvalid = false
			}
		}

		if this.exit {
			this.isvalid = false
			return
		}
	}

}

//根据设计，这个aoc client必须允许重连
func (this *KeepAliveWebsocketClient) connect(timeout int) error {
	if this.isvalid {
		return nil
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	if this.url == "" {
		return errors.New("websocket url is nil or empty")
	}

	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Millisecond * time.Duration(timeout)
	c, _, err := dialer.Dial(this.url, nil)
	if err != nil {
		this.url = ""
		return err
	}

	this.client = c
	this.isvalid = true
	return nil
}

func (this *KeepAliveWebsocketClient) Close() {
	this.exit = true

	this.locker.Lock()
	this.isvalid = false
	defer this.locker.Unlock()

	if this.client != nil {
		//_ =this.client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		_ = this.client.Close()
	}
}

func (this *KeepAliveWebsocketClient) Send(msg []byte) error {
	this.locker.Lock()

	if !this.isvalid {
		this.locker.Unlock()
		return errors.New("websocket client valid is not true")
	}

	if this.client == nil {
		this.locker.Unlock()
		return errors.New("websocket client not inited")
	}

	err := this.client.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		this.locker.Unlock()
		return err
	}

	this.locker.Unlock()
	return nil
}

func (this *KeepAliveWebsocketClient) pingPongCheck(timeout int) error {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.isvalid {
		return errors.New("websocket client valid is not true")
	}

	if this.client == nil {
		return errors.New("websocket client not inited")
	}

	err := this.client.WriteMessage(websocket.PingMessage, nil)
	if err != nil {
		return err
	}

	this.client.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(timeout)))
	mgsType, _, err := this.client.ReadMessage()
	if err != nil {
		return err
	}
	if mgsType != websocket.PongMessage {
		return errors.New("ws recv message type is not pong")
	}

	return nil
}
