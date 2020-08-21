/**
 * @Author: zhangyw
 * @Description:
 * @File:  Repo
 * @Date: 2020/5/9 10:09
 */

package srvDiscover

import (
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/xukgo/gsaber/utils/randomUtil"
	"github.com/xukgo/gsaber/utils/stringUtil"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

//const DEFAULT_CONN_TIMEOUT = 1500

type Repo struct {
	locker sync.RWMutex
	config *ConfRoot
	client *clientv3.Client //etcd客户端

	subsNodeCache map[string]*SubSrvNodeList

	subLicResultInfo *SubLicResultInfo
	licLocker        sync.RWMutex
	licPrivkey       string
	licWatchFunc     func(*LicResultInfo)
}

func (this *Repo) InitFromPath(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	srvConf := new(ConfRoot)
	err = srvConf.FillWithXml(content)
	if err != nil {
		return err
	}

	this.config = srvConf

	this.client, err = clientv3.New(clientv3.Config{
		Endpoints:   srvConf.Endpoints,
		DialTimeout: time.Duration(srvConf.Timeout) * time.Second,
	})
	return nil
}

func (this *Repo) StartRegister(beforeRegisterFunc BeforeRegisterFunc) error {
	if this.config == nil {
		return fmt.Errorf("register conf is nil")
	}

	registerOp := this.config.GetRegisterOptionFuncs()
	if beforeRegisterFunc != nil {
		registerOp = append(registerOp, WithBeforeRegister(beforeRegisterFunc))
	}

	srvInfo, err := this.config.GetRegisterModule()
	if err != nil {
		return err
	}

	go this.Register(srvInfo, registerOp...)
	return nil
}

func (this *Repo) StartSubscribe() error {
	if this.config == nil {
		return fmt.Errorf("register conf is nil")
	}

	subBasicInfos, err := this.config.GetSubscribeBasicInfos()
	if err != nil {
		return err
	}

	this.initSubsNodeCache(subBasicInfos)
	err = this.SubScribe(subBasicInfos)
	return err
}

func (this *Repo) GetLocalRegisterInfo() *RegisterConf {
	conf := this.config
	return conf.RegisterConf
}

func (this *Repo) GetSubsNames() []string {
	subsconf := this.config.SubScribeConf
	if subsconf == nil {
		return nil
	}
	if len(subsconf.Services) == 0 {
		return nil
	}

	arr := make([]string, 0, len(subsconf.Services))
	for idx := range subsconf.Services {
		arr = append(arr, subsconf.Services[idx].Name)
	}
	return arr
}
func (this *Repo) GetServiceByName(name string) []*RegisterInfo {
	this.locker.RLock()
	defer this.locker.RUnlock()

	var srvInfos []*RegisterInfo = nil
	for srvName, srvNodeList := range this.subsNodeCache {
		if stringUtil.CompareIgnoreCase(srvName, name) {
			srvInfos = make([]*RegisterInfo, 0, len(srvNodeList.NodeInfos))
			for n := range srvNodeList.NodeInfos {
				srvInfos = append(srvInfos, srvNodeList.NodeInfos[n].RegInfo.DeepClone())
			}
			break
		}
	}
	return srvInfos
}

func (this *Repo) GetRandomServiceArray(svcName string) []*RegisterInfo {
	infos := this.GetServiceByName(svcName)
	if len(infos) == 0 {
		return nil
	}

	randomSortSlice(infos)
	return infos
}

func (this *Repo) GetRandomServiceByName(name string) *RegisterInfo {
	this.locker.RLock()
	defer this.locker.RUnlock()

	for srvName, srvNodeList := range this.subsNodeCache {
		count := len(srvNodeList.NodeInfos)
		if stringUtil.CompareIgnoreCase(srvName, name) {
			if count == 0 {
				return nil
			} else if count == 1 {
				return srvNodeList.NodeInfos[0].RegInfo.DeepClone()
			} else {
				idx := randomUtil.NewInt32(0, int32(count))
				return srvNodeList.NodeInfos[idx].RegInfo.DeepClone()
			}
		}
	}
	return nil
}

func (this *Repo) GetServiceByNameAndNodeId(name string, id string) *RegisterInfo {
	this.locker.RLock()
	defer this.locker.RUnlock()

	for srvName, srvNodeList := range this.subsNodeCache {
		if srvName != name {
			continue
		}
		nodes := srvNodeList.NodeInfos
		for n := range nodes {
			if nodes[n].RegInfo.Global.NodeId == id {
				return srvNodeList.NodeInfos[n].RegInfo.DeepClone()
			}
		}
	}
	return nil
}

func (this *Repo) initSubsNodeCache(subSrvInfos []SubBasicInfo) {
	serviceCount := len(subSrvInfos)
	if serviceCount <= 0 {
		return
	}

	this.subsNodeCache = make(map[string]*SubSrvNodeList)
	for m := 0; m < serviceCount; m++ {
		srvNodeList := new(SubSrvNodeList)
		srvNodeList.SubBasicInfo = *NewSubSrvBasicInfo(subSrvInfos[m].Name, subSrvInfos[m].Version, subSrvInfos[m].Namespace)
		srvNodeList.NodeInfos = make([]*SrvNodeInfo, 0, 1)
		this.subsNodeCache[subSrvInfos[m].Name] = srvNodeList
	}
}

//随机打乱数组
func randomSortSlice(arr []*RegisterInfo) {
	if len(arr) <= 0 || len(arr) == 1 {
		return
	}

	for i := len(arr) - 1; i > 0; i-- {
		num := randomUtil.NewInt32(0, int32(i+1))
		arr[i], arr[num] = arr[num], arr[i]
	}
}

//func GetSrvDiscover() *Repo {
//	return srvDiscoverInstance
//}
//
//func GetSrvDiscoverConf() *ConfRoot {
//	return &srvDiscoverConf
//}

//func (this *ServiceDiscovery) TriggerRegister() {
//	this.registerHupChan <- true
//}
//

//func (this *Repo) GetConfig() *ConfRoot {
//	return this.config
//}

//func NewSrvDiscover(endpoints []string, options ...SdOption) (*Repo, error) {
//	if len(endpoints) == 0 {
//		return nil, fmt.Errorf("endpoints addrs is empty")
//	}
//
//	serviceDiscovery := &Repo{
//		Endpoints: endpoints,
//	}
//	serviceDiscovery.Timeout = DEFAULT_CONN_TIMEOUT * time.Millisecond
//
//	for _, op := range options {
//		op(serviceDiscovery)
//	}
//
//	var err error
//	serviceDiscovery.client, err = clientv3.New(clientv3.Config{
//		Endpoints:   endpoints,
//		DialTimeout: serviceDiscovery.Timeout,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	serviceDiscovery.subsNodeCache = make(map[string]*SubSrvNodeList)
//	serviceDiscovery.locker = &sync.RWMutex{}
//
//	return serviceDiscovery, nil
//}
