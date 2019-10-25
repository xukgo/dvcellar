package muplxWsPool

import (
	"github.com/xukgo/dvcellar/comline"
	"errors"
	"math/rand"
	"net/url"
	"time"
)

type comlineWrapper struct {
	getCount uint64
	conn     *comLine.KeepAliveWebsocketClient
}

type Pool struct {
	liveCount    int
	timeout      int
	url          string
	comlineGroup []*comlineWrapper
}

func NewPool(wsUrl string, connectTimeout int, livecnt int) (*Pool, error) {
	pool := new(Pool)
	pool.url = wsUrl
	pool.liveCount = livecnt
	pool.timeout = connectTimeout

	if livecnt <= 0 {
		return nil, errors.New("ws pool live count must > 0")
	}

	_, err := url.Parse(wsUrl)
	if err != nil {
		return nil, errors.New("ws url is not a valid format")
	}
	return pool, nil
}

func (this *Pool) StartKeepAlive() {
	for i := 0; i < this.liveCount; i++ {
		comline := comLine.NewKeepAliveWebsocketClient(this.url)
		go comline.StartKeepAlive(this.timeout)

		wrapper := &comlineWrapper{
			getCount: 0,
			conn:     comline,
		}
		this.comlineGroup = append(this.comlineGroup, wrapper)
	}
}

func (this *Pool) GetUrl() string {
	return this.url
}
func (this *Pool) Get() *comLine.KeepAliveWebsocketClient {
	var validIndexArr []int
	comlineLen := len(this.comlineGroup)
	for i := 0; i < comlineLen; i++ {
		if this.comlineGroup[i].conn.IsValid() {
			validIndexArr = append(validIndexArr, i)
		}
	}

	if validIndexArr == nil || len(validIndexArr) == 0 {
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	selectIndex := validIndexArr[r.Intn(len(validIndexArr))]
	this.comlineGroup[selectIndex].getCount++
	return this.comlineGroup[selectIndex].conn
}

func (this *Pool) Close() {
	for _, wrapper := range this.comlineGroup {
		wrapper.conn.Close()
	}
}
