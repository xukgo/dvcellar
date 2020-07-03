/**
 * @Author: xuk
 * @Description:
 * @File:  Conf
 * @Date: 2020/6/3 9:25
 */

package srvDiscover

import (
	"encoding/xml"
	"fmt"
	"github.com/xukgo/gsaber/utils/netUtil"
	"strings"
	"time"
)

/*
go mod edit -replace github.com/coreos/bbolt@v1.3.4=go.etcd.io/bbolt@v1.3.4
go mod edit -replace google.golang.org/grpc@v1.29.1=google.golang.org/grpc@v1.26.0
*/
type ConfRoot struct {
	XMLName       xml.Name
	Timeout       int            `xml:"Timeout"`        //etcd连接超时时间,单秒秒
	Endpoints     []string       `xml:"Endpoints>Addr"` //etcd服务器地址, 172.16.0.212:2379
	RegisterConf  *RegisterConf  `xml:"Register"`
	SubScribeConf *SubscribeConf `xml:"Subscribe"`
}

type RegisterConf struct {
	Interval  int                     `xml:"Interval"`  //注册间隔, 单位秒, 默认值为2
	TTL       int                     `xml:"TTL"`       //注册服务的TimeToLive, 单位秒,默认值为6
	Namespace string                  `xml:"Namespace"` //注册Key的namespace, 默认为voice, /registry/namespace/..
	Global    RegisterGlobalConf      `xml:"Global"`
	SvcInfos  []RegisterSvcDefineConf `xml:"SvcInfos>Svc"`
	//PrivateMap []SrvRegisterPrivateConf `xml:"PrivateMap>Private"`
}

type RegisterGlobalConf struct {
	Name     string `xml:"Name"`
	State    string `xml:"State"`
	NodeId   string `xml:"NodeId"`
	Version  string `xml:"Version"`
	IPString string `xml:"IP"`
	IP       string `xml:"-"`
}

type RegisterSvcDefineConf struct {
	Name string `xml:"name,attr" json:"name"`
	Port int    `xml:"port,attr" json:"port"`
}

func (this *RegisterSvcDefineConf) DeepClone() *RegisterSvcDefineConf {
	model := new(RegisterSvcDefineConf)
	model.Name = this.Name
	model.Port = this.Port
	return model
}

type SubscribeSrvConf struct {
	Namespace string `xml:"Namespace"`
	Name      string `xml:"Name"`
	Version   string `xml:"Version"`
}

type SubscribeConf struct {
	Services []SubscribeSrvConf `xml:"Service"`
}

func (this *ConfRoot) FillWithXml(data []byte) error {
	err := xml.Unmarshal(data, this)
	if err != nil {
		return err
	}

	//反序列化后的处理
	if this.Timeout <= 0 {
		this.Timeout = 2
	}

	this.RegisterConf.Namespace = strings.TrimSpace(this.RegisterConf.Namespace)
	if len(this.RegisterConf.Namespace) == 0 {
		this.RegisterConf.Namespace = defaultRegisterOption.Namespace
	}
	if this.RegisterConf.TTL == 0 {
		this.RegisterConf.TTL = int(defaultRegisterOption.TTLSec)
	}
	if this.RegisterConf.Interval == 0 {
		this.RegisterConf.Interval = int(defaultRegisterOption.Interval / time.Second)
	}

	if this.SubScribeConf != nil {
		for idx := range this.SubScribeConf.Services {
			if len(this.SubScribeConf.Services[idx].Namespace) == 0 {
				this.SubScribeConf.Services[idx].Namespace = defaultSubscribeOption.Namespace
			}
		}
	}
	ip, err := convertRegisteIP(this.RegisterConf.Global.IPString)
	if err != nil {
		return err
	}
	this.RegisterConf.Global.IP = ip
	return err
}

func convertRegisteIP(ipString string) (string, error) {
	arr := strings.Split(ipString, ":")
	var filterArr []string
	if len(arr) == 1 {
		filterArr = nil
	} else {
		filterArr = strings.Split(arr[1], "|")
	}

	ipArr, err := netUtil.GetIPv4(arr[0], filterArr)
	if err != nil {
		return "", err
	}
	if len(ipArr) == 0 {
		return "", fmt.Errorf("no ip found")
	}
	return ipArr[0], nil
}

func (this *ConfRoot) GetRegisterOptionFuncs() []RegisterOptionFunc {
	if this.RegisterConf == nil {
		return nil
	}

	register := this.RegisterConf
	registerOp := make([]RegisterOptionFunc, 0, 4)
	registerOp = append(registerOp, WithTTL(int64(register.TTL)))
	registerOp = append(registerOp, WithRegisterNamespace(register.Namespace))
	registerOp = append(registerOp, WithRegisterInterval(time.Duration(register.Interval)*time.Second))
	registerOp = append(registerOp, WithRegisterConnTimeout(time.Duration(this.Timeout)*time.Second))
	return registerOp
}

func (this *ConfRoot) GetRegisterModule() (*RegisterInfo, error) {
	if this.RegisterConf == nil {
		return nil, fmt.Errorf("register conf is nil")
	}

	register := this.RegisterConf
	if len(register.Global.Name) == 0 {
		return nil, fmt.Errorf("register global.name is empty")
	}
	if len(register.Global.NodeId) == 0 {
		return nil, fmt.Errorf("register global.nodeId is empty")
	}
	if len(register.Global.Version) == 0 {
		return nil, fmt.Errorf("register global.version is empty")
	}
	if len(register.Global.IP) == 0 {
		return nil, fmt.Errorf("register global.ip is empty")
	}

	srvInfo := new(RegisterInfo)
	srvInfo.Global.Name = register.Global.Name
	srvInfo.Global.NodeId = register.Global.NodeId
	srvInfo.Global.Version = register.Global.Version
	srvInfo.Global.IP = register.Global.IP
	srvInfo.Global.State = register.Global.State

	srvInfo.SvcInfos = register.SvcInfos
	//if len(register.SvcInfos) > 0 {
	//	portMap := register.SvcInfos
	//	srvInfo.Port = make(map[string]int)
	//	for i := range portMap {
	//		srvInfo.Port[portMap[i].Name] = portMap[i].Port
	//	}
	//} else {
	//	srvInfo.Port = nil
	//}

	return srvInfo, nil
}

func (this *ConfRoot) GetSubscribeBasicInfos() ([]SubBasicInfo, error) {
	if this.SubScribeConf == nil {
		return nil, fmt.Errorf("subscribe conf is nil")
	}
	subSrvs := this.SubScribeConf.Services
	infos := make([]SubBasicInfo, 0, len(subSrvs))
	for i := range subSrvs {
		infos = append(infos, *NewSubSrvBasicInfo(subSrvs[i].Name, subSrvs[i].Version, subSrvs[i].Namespace))
	}

	return infos, nil
}
