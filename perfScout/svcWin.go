// +build windows

/*
@Time : 2019/9/24 20:34
@Author : Hermes
@File : svcWin
@Description:
*/
package perfScout

import "os"

var singleton *perfInfoMan

func Start() {
	singleton = new(perfInfoMan)
	go loopUpdateMonitorInfo()
}

func loopUpdateMonitorInfo() {
	hostName, _ := os.Hostname()
	singleton.setHostName(hostName)
}
