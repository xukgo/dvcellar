package svcDiscovery

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/xukgo/gsaber/utils/netUtil"
	"strconv"
	"time"
)

type registerResponse struct {
	Result registerResult `json:"result"`
	//Service string        `json:"service"`
}
type registerResult struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (this *registerResponse) fillWithJson(gson string) error {
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(gson), this)
	if err != nil {
		return err
	}
	return nil
}

func (this *serviceDiscoveryInfo) formatRegisterUrl() string {
	return fmt.Sprintf("http://%s:%d", this.IP, this.Port)
}

func (this *serviceDiscoveryInfo) Start(getRegisterInfoFunc func() (RegisterSystemInfo, interface{}), localInfo LocalServiceDefine, timeout int, interval int) {
	url := this.formatRegisterUrl()
	for {
		sysInfo, privInfo := getRegisterInfoFunc()
		regInfo := registerInfo{}
		regInfo.Command = "ServiceRegister"
		regInfo.System = sysInfo
		regInfo.Private = privInfo
		regInfo.Global = createGloabalInfo(localInfo)

		gson, _ := regInfo.toJson()

		startTime := time.Now()
		resp, err := registerPost(url, gson, timeout)
		endTime := time.Now()
		elapse := int((endTime.UnixNano() - startTime.UnixNano()) / 1000000)

		if err != nil || resp.Result.Code != 0 {
			time.Sleep(time.Second * 1)
			continue
		}
		if elapse < interval {
			time.Sleep(time.Millisecond * time.Duration(interval-elapse))
		}
	}
}

func createGloabalInfo(info LocalServiceDefine) registerGlobalInfo {
	model := registerGlobalInfo{
		Name:      info.Name,
		HostName:  "none",
		State:     "online",
		NodeId:    info.NodeId,
		Version:   info.Version,
		Timestamp: fmt.Sprintf("%d", time.Now().UnixNano()/1000000),
	}

	model.HttpAddr = info.HttpUrl
	model.WebAddr = info.LocalIP
	model.WebPort = info.WebPort
	return model
}

func registerPost(url string, gson string, timeout int) (*registerResponse, error) {
	httpResponse := netUtil.HttpPostJson(url, gson, timeout, true)
	if httpResponse.Error != nil {
		return nil, httpResponse.Error
	}
	if httpResponse.StatusCode != 200 {
		return nil, errors.New("http code:" + strconv.Itoa(httpResponse.StatusCode))
	}

	resp := new(registerResponse)
	err := resp.fillWithJson(string(httpResponse.Data))
	if err != nil {
		return nil, err
	}

	return resp, nil
}
