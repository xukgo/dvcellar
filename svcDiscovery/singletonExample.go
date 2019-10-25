package svcDiscovery

//import (
//	"github.com/xukgo/gsaber/utils/fileUtil"
//)
//
//var singleton = new(xvoiceSvcDisc.ServieDiscoveryRepo)
//
//type PrivateInfo struct {
//	SessionCount int `json:"sessionCount"`
//	HttpConn     int `json:"httpConn"`
//}
//
//func GetSingleton() *xvoiceSvcDisc.ServieDiscoveryRepo {
//	return singleton
//}
//func StartServiceDiscovery() error {
//	fileUrl := fileUtil.GetAbsUrl("conf/ServiceDiscovery.xml")
//	err := singleton.InitConf(fileUrl)
//	if err != nil {
//		return err
//	}
//
//	singleton.StartRegister(getRegisterInfoFunc)
//	singleton.StartSubsvc()
//
//	return nil
//}
//
//func getRegisterInfoFunc() (xvoiceSvcDisc.RegisterSystemInfo, interface{}) {
//	sysInfo := xvoiceSvcDisc.RegisterSystemInfo{}
//	monitorInfo := systemMonitor.GetSystemInfo()
//	sysInfo.Cpu = int(monitorInfo.CpuLoadPercent)
//	sysInfo.IO = int(monitorInfo.WriteIOSpeed)
//	sysInfo.Disk = 0
//	sysInfo.Memmory = int(monitorInfo.MemUse)
//	sysInfo.Tcp = 0
//	sysInfo.Socket = int(monitorInfo.FdCount)
//
//	privInfo := &PrivateInfo{}
//
//	return sysInfo, privInfo
//}
