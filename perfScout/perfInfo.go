/*
@Time : 2019/9/24 20:32
@Author : Hermes
@File : systemInfo
@Description:
*/
package perfScout

import "sync"

type perfInfoMan struct {
	PerfInfo
	locker sync.RWMutex
}

type PerfInfo struct {
	HostName       string
	CpuLoadPercent float64
	MemUse         int `unit:M`
	FdCount        int
	SocketCount    int
	WriteIOSpeed   float64
}

func GetPerfInfo() PerfInfo {
	singleton.locker.RLock()
	resInfo := singleton.PerfInfo
	singleton.locker.RUnlock()
	return resInfo
}

func (this *perfInfoMan) setHostName(name string) {
	this.locker.Lock()
	this.HostName = name
	this.locker.Unlock()
}
