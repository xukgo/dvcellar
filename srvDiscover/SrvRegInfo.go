/**
 * @Author: zhangyw
 * @Description:
 * @File:  RegisterInfo
 * @Date: 2020/5/11 13:34
 */

package srvDiscover

import (
	"crypto/md5"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"time"
)

type RegisterGlobalInfo struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	NodeId    string `json:"nodeId"`
	Version   string `json:"version"`
	IP        string `json:"ip"`
	Timestamp string `json:"timestamp"`
}

func (this *RegisterGlobalInfo) RefreshTimestamp(dt time.Time) {
	this.Timestamp = fmt.Sprintf("%d", dt.UnixNano()/int64(time.Millisecond))
}

type RegisterProfileInfo struct {
	Cpu    int `json:"cpu"`
	IO     int `json:"io"`
	Disk   int `json:"disk"`
	Memory int `json:"memory"`
	Socket int `json:"socket"`
}

type RegisterInfo struct {
	Global  RegisterGlobalInfo  `json:"global"`
	Profile RegisterProfileInfo `json:"profile"`
	Port    map[string]int      `json:"port"`
	Private map[string]string   `json:"private"`
}

func (this RegisterInfo) FormatRegisterKey(namespace string) string {
	key := fmt.Sprintf("registry.%s.%s.%s", namespace, this.GetServiceName(), this.UniqueId())
	return key
}

func (this *RegisterInfo) Serialize() []byte {
	gson, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(this)
	return gson
}

func (this *RegisterInfo) Deserialize(data []byte) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, this)
}

func (this *RegisterInfo) GetServiceName() string {
	return this.Global.Name
}

func (this *RegisterInfo) UniqueId() string {
	md5Str := fmt.Sprintf("%x", md5.Sum([]byte(this.Global.IP+this.Global.NodeId)))
	return md5Str
}

func (this *RegisterInfo) DeepClone() *RegisterInfo {
	model := new(RegisterInfo)
	model.Global = this.Global
	model.Profile = this.Profile

	if this.Port == nil {
		model.Port = nil
	} else {
		model.Port = make(map[string]int)
		for key, value := range this.Port {
			model.Port[key] = value
		}
	}

	if this.Private == nil {
		model.Private = nil
	} else {
		model.Private = make(map[string]string)
		for key, value := range this.Private {
			model.Private[key] = value
		}
	}
	return model
}

//func (this *RegisterInfo) MakeEmpty() {
//	*this = RegisterInfo{}
//}
