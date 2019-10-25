package svcDiscovery

import (
	"fmt"
	"github.com/json-iterator/go"
)

type serviceDiscoveryInfo struct {
	IP   string
	Port int
}

func newServiceDiscoveryInfo(ip string, port int) *serviceDiscoveryInfo {
	model := new(serviceDiscoveryInfo)
	model.IP = ip
	model.Port = port
	return model
}

type registerInfo struct {
	Command string             `json:"command"`
	Global  registerGlobalInfo `json:"global"`
	System  RegisterSystemInfo `json:"system"`
	Private interface{}        `json:"private,omiempty"`
}

type registerGlobalInfo struct {
	Name     string `json:"name"`
	NodeId   string `json:"nodeId"`
	Version  string `json:"version"`
	State    string `json:"state"`
	HostName string `json:"hostName"`

	HttpAddr string `json:"httpAddr"`
	WebAddr  string `json:"webAddr"`
	WebPort  int    `json:"webPort"`

	Timestamp string `json:"timestamp"`
}

type RegisterSystemInfo struct {
	Cpu     int `json:"cpu"`
	IO      int `json:"io"`
	Disk    int `json:"disk"`
	Memmory int `json:"mem"`
	Tcp     int `json:"tcp"`
	Socket  int `json:"socket"`
}

func (this *registerInfo) toJson() (string, error) {
	gson, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(this)
	if err != nil {
		fmt.Println("registerInfo Marshal json string error")
		return "", err
	}

	return string(gson), nil
}
