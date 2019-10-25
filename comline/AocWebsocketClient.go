//和switch对接的时候由于nopoll的问题，用gorilla作为client连接有问题，切换到x/net库
package comLine

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type AocWebsocket struct {
	client  *websocket.Conn
	url     string
	locker  sync.Mutex
	isvalid bool
}

func NewAocWebsocket(inUrl string) *AocWebsocket {
	return &AocWebsocket{
		url:     inUrl,
		isvalid: false,
	}
}

//根据设计，这个aoc client必须允许重连
func (this *AocWebsocket) Connect(timeout int) error {
	if this.isvalid {
		return nil
	}

	this.locker.Lock()
	if this.isvalid {
		this.locker.Unlock()
		return nil
	}

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

func (this *AocWebsocket) Close() {
	this.locker.Lock()
	this.isvalid = false
	defer this.locker.Unlock()

	if this.client == nil {
		//_ =this.client.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		_ = this.client.Close()
	}

	this.client = nil
}

func (this *AocWebsocket) Send(msg []byte) error {
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
