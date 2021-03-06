package svcDiscovery

import (
	"fmt"
	"github.com/xukgo/gsaber/utils/randomUtil"
	"os"
	"sync"
)

type ServieDiscoveryRepo struct {
	confLocker sync.RWMutex
	sdConf     *confXmlModel
	register   *serviceDiscoveryInfo
	watcher    *subSvcWatcher
}

func (this *ServieDiscoveryRepo) InitConf(fileUrl string) error {
	xmlContent, err := os.ReadFile(fileUrl)
	if err != nil {
		errMsg := fmt.Sprintf("ReadFile %s error", fileUrl)
		fmt.Println("ServieDiscoveryRepo InitConf" + errMsg)
		return err
	}

	err = this.updateFromContent(xmlContent)
	if err != nil {
		errMsg := fmt.Sprintf("ServieDiscoveryRepo updateFromContent error:" + err.Error())
		fmt.Println(errMsg)
		return err
	}
	fmt.Println("ServieDiscoveryRepo InitConf success")
	return nil
}

func (this *ServieDiscoveryRepo) AddObserver(observer Observer) {
	if this.watcher != nil {
		this.watcher.addObserver(observer)
	}
}

func (this *ServieDiscoveryRepo) StartRegister(getRegisterInfoFunc func() (RegisterSystemInfo, interface{})) {
	conf := this.sdConf
	if conf == nil {
		return
	}

	this.register = newServiceDiscoveryInfo(conf.SdDomain.IP, conf.SdDomain.Port)
	go this.register.Start(getRegisterInfoFunc, conf.LocalSvc, conf.SdDomain.Timeout, conf.LocalSvc.UpdateInterval)
}

func (this *ServieDiscoveryRepo) StartSubsvc() {
	conf := this.sdConf
	this.watcher = newSubSvcWatcher(this, conf.SdDomain.IP, conf.SdDomain.Port)

	var infos []subSvcReuqestGloablInfo
	for _, item := range conf.SubSvcs.SubServices {
		infos = append(infos, *newSubSvcReuqestGloablInfo(item.Name, item.Version))
	}
	if len(infos) == 0 {
		return
	}
	go this.watcher.Start(infos, conf.SdDomain.Timeout, conf.SubSvcs.SubsInterval)
}

func (this *ServieDiscoveryRepo) updateFromContent(content []byte) error {
	this.confLocker.Lock()
	defer this.confLocker.Unlock()

	model := &confXmlModel{}
	err := model.fillWithXml(content)
	if err != nil {
		return err
	}

	this.sdConf = model
	return nil
}

func (this *ServieDiscoveryRepo) GetLocalServiceDefine() LocalServiceDefine {
	this.confLocker.RLock()
	info := (*this.sdConf).LocalSvc
	this.confLocker.RUnlock()
	return info
}

func (this *ServieDiscoveryRepo) GetSubSvcNames() []string {
	this.confLocker.RLock()
	conf := this.sdConf
	var infos []string
	for _, item := range conf.SubSvcs.SubServices {
		if len(item.Name) > 0 {
			infos = append(infos, item.Name)
		}
	}
	this.confLocker.RUnlock()
	return infos
}

func (this *ServieDiscoveryRepo) GetServiceInfos(name string, isRandSort bool) []SvcInfo {
	if this.watcher == nil {
		return nil
	}
	arr := this.watcher.getArrayByName(name)
	if isRandSort {
		randomSortSlice(arr)
	}
	return arr
}

//随机打乱数组
func randomSortSlice(arr []SvcInfo) {
	if len(arr) <= 0 || len(arr) == 1 {
		return
	}

	for i := len(arr) - 1; i > 0; i-- {
		num := randomUtil.NewInt32(0, int32(i+1))
		arr[i], arr[num] = arr[num], arr[i]
	}
}
