package softTime

import (
	"sync/atomic"
	"time"
)

var dtNowTs int64 = 0

func Start(interval time.Duration) {
	ticker1 := time.NewTicker(interval)
	go func(t *time.Ticker) {
		for {
			<-t.C
			t := time.Now().UnixNano()
			atomic.StoreInt64(&dtNowTs, t)
		}
	}(ticker1)
}

func NowUnixNano() int64 {
	t := atomic.LoadInt64(&dtNowTs)
	if t == 0 {
		return time.Now().UnixNano()
	}
	return t
}

func Now() time.Time {
	if dtNowTs == 0 {
		return time.Now()
	}
	return time.Unix(0, dtNowTs)
}
