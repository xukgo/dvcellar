package svcDiscovery

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/xukgo/gsaber/utils/netUtil"
	"strconv"
	"strings"
	"sync"
	"time"
)

type subSvcReuqest struct {
	Command    string                  `json:"command"`
	GlobalInfo subSvcReuqestGloablInfo `json:"global"`
}

type subSvcReuqestGloablInfo struct {
	Name    string `json:"name"`
	Version string `json:"-"`
}

func newSubSvcReuqestGloablInfo(name, version string) *subSvcReuqestGloablInfo {
	model := new(subSvcReuqestGloablInfo)
	model.Name = name
	model.Version = version
	return model
}

func (this *subSvcReuqest) toJson() (string, error) {
	gson, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(this)
	if err != nil {
		fmt.Println("subSvcReuqest Marshal json string error")
		return "", err
	}

	return string(gson), nil
}

type subSvcResponse struct {
	Notify   string       `json:"notify"`
	Result   subSvcResult `json:"result"`
	SvcInfos []SvcInfo    `json:"subject"`
}

type subSvcResult struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type SvcInfo struct {
	GlobalInfo SvcGlobalInfo `json:"global"`
}

type SvcGlobalInfo struct {
	Name     string `json:"name"`
	State    string `json:"state"`
	NodeId   string `json:"nodeId"`
	Version  string `json:"version"`
	HttpAddr string `json:"httpAddr"`
	WebAddr  string `json:"webAddr"`
	WebPort  int    `json:"webPort"`

	Timestamp string `json:"timestamp"`
}

func (this *subSvcResponse) fillWithJson(gson string) error {
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(gson), this)
	if err != nil {
		return err
	}
	return nil
}

type subSvcWatcher struct {
	repo     *ServieDiscoveryRepo
	SdIP     string
	SdPort   int
	svcArray []SvcInfo
	rwLocker sync.RWMutex
	//lastProvidIndex int //从1开始，默认初始化是0
	observerArray  []Observer
	observerLocker sync.Mutex
}

func newSubSvcWatcher(repo *ServieDiscoveryRepo, ip string, port int) *subSvcWatcher {
	model := new(subSvcWatcher)
	model.repo = repo
	model.SdIP = ip
	model.SdPort = port
	return model
}

func (this *subSvcWatcher) addObserver(observer Observer) {
	this.observerLocker.Lock()
	this.observerArray = append(this.observerArray, observer)
	this.observerLocker.Unlock()
}

func (this *subSvcWatcher) Start(infos []subSvcReuqestGloablInfo, timeout int, interval int) {
	url := this.formatGetServiceUrl()
	for {
		startTime := time.Now()

		isReponseValid := false
		for idx := range infos {
			svcRes, err := getService(infos[idx], url, timeout)
			if err != nil {
				continue
			}

			if svcRes.Result.Code != 0 {
				continue
			}

			isReponseValid = true
			if svcRes.SvcInfos == nil || len(svcRes.SvcInfos) == 0 {
				this.deleteSvcs(infos[idx].Name)
			} else {
				this.updateSvcs(svcRes.SvcInfos, infos[idx])
			}
		}

		if isReponseValid {
			this.observerLocker.Lock()
			for _, obs := range this.observerArray {
				obs.UpdateFromSvcDisc(this.repo)
			}
			this.observerLocker.Unlock()
		}

		//fmt.Println(this.svcArray)
		endTime := time.Now()
		elapse := int((endTime.UnixNano() - startTime.UnixNano()) / 1000000)
		if elapse < interval {
			time.Sleep(time.Millisecond * time.Duration(interval-elapse))
		}
	}
}

func (this *subSvcWatcher) formatGetServiceUrl() string {
	return fmt.Sprintf("http://%s:%d", this.SdIP, this.SdPort)
}

func (this *subSvcWatcher) updateSvcs(infos []SvcInfo, destInfo subSvcReuqestGloablInfo) {
	if infos == nil || len(infos) == 0 {
		return
	}

	this.rwLocker.Lock()
	this.svcArray = deleteSvcSliceByName(this.svcArray, destInfo.Name)
	for _, item := range infos {
		if strings.Contains(item.GlobalInfo.Version, destInfo.Version) {
			this.svcArray = append(this.svcArray, item)
		}
	}
	this.rwLocker.Unlock()
}

func (this *subSvcWatcher) deleteSvcs(name string) {
	this.rwLocker.Lock()
	this.svcArray = deleteSvcSliceByName(this.svcArray, name)
	this.rwLocker.Unlock()
}

func deleteSvcSliceByName(arr []SvcInfo, name string) []SvcInfo {
	j := 0
	for _, val := range arr {
		if strings.ToLower(val.GlobalInfo.Name) != strings.ToLower(name) {
			arr[j] = val
			j++
		}
	}
	return arr[:j]
}

func getService(info subSvcReuqestGloablInfo, url string, timeout int) (*subSvcResponse, error) {
	requestModel := new(subSvcReuqest)
	requestModel.Command = "GetService"
	requestModel.GlobalInfo = info

	gson, _ := requestModel.toJson()
	httpResponse := netUtil.HttpPostJson(url, gson, timeout,true)
	if httpResponse.Error != nil {
		return nil, httpResponse.Error
	}
	if httpResponse.StatusCode != 200 {
		return nil, errors.New("http code:" + strconv.Itoa(httpResponse.StatusCode))
	}

	subsRes := new(subSvcResponse)
	err := subsRes.fillWithJson(string(httpResponse.Data))
	if err != nil {
		return nil, err
	}

	return subsRes, nil
}

func (this *subSvcWatcher) getArrayByName(name string) []SvcInfo {
	this.rwLocker.RLock()
	var resArr = make([]SvcInfo,0, len(this.svcArray)/2)
	for _, item := range this.svcArray {
		if checkSvcInfoMatchNameAndOnline(item, name) {
			resArr = append(resArr, item)
		}
	}
	this.rwLocker.RUnlock()
	return resArr
}

//获取一个服务，轮流获取
//func (this *subSvcWatcher) getByName(name string) *SvcInfo {
//	this.rwLocker.RLock()
//	var resInfo SvcInfo
//
//	for i := this.lastProvidIndex; i <= len(this.svcArray); i++ {
//		if checkSvcInfoMatchNameAndOnline(this.svcArray[i], name) {
//			resInfo = this.svcArray[i]
//			this.lastProvidIndex = i + 1
//			this.rwLocker.RUnlock()
//			return &resInfo
//		}
//	}
//	for i := 0; i <= len(this.svcArray); i++ {
//		if checkSvcInfoMatchNameAndOnline(this.svcArray[i], name) {
//			resInfo = this.svcArray[i]
//			this.lastProvidIndex = i + 1
//			this.rwLocker.RUnlock()
//			return &resInfo
//		}
//	}
//	this.rwLocker.RUnlock()
//	return nil
//}

func checkSvcInfoMatchNameAndOnline(info SvcInfo, name string) bool {
	if strings.ToLower(info.GlobalInfo.State) != "online" {
		return false
	}
	if strings.ToLower(info.GlobalInfo.Name) != strings.ToLower(name) {
		return false
	}
	//if len(version) > 0 && !strings.Contains(item.GlobalInfo.Version, version) {
	//	return false
	//}

	return true
}
