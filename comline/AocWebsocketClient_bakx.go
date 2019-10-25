package comLine

//import (
//	"errors"
//	"golang.org/x/net/websocket"
//	"sync"
//)
//
//type AocWebsocket struct {
//	client *websocket.Conn
//	url    string
//	locker sync.Mutex
//}
//
//func NewAocWebsocket(inUrl string) *AocWebsocket {
//	return &AocWebsocket{
//		url: inUrl,
//	}
//}
//
//func (this *AocWebsocket) Init() error {
//	if this.url == "" {
//		return errors.New("AocWebsocket url is nil or empty")
//	}
//
//	origin := "http://localhost/"
//	c, err := websocket.Dial(this.url, "", origin)
//	if err != nil {
//		return err
//	}
//	this.client = c
//	return nil
//}
//
//func (this *AocWebsocket) Close() error {
//	if this.client == nil {
//		return nil
//	}
//
//	this.client.Close()
//
//	this.client = nil
//	return nil
//}
//
//func (this *AocWebsocket) Send(msg []byte) error {
//	if this.client == nil {
//		return errors.New("websocket client not inited")
//	}
//
//	this.locker.Lock()
//	_, err := this.client.Write(msg)
//	if err != nil {
//		this.locker.Unlock()
//		return err
//	}
//
//	this.locker.Unlock()
//	return nil
//}
