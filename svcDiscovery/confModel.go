package svcDiscovery

import "encoding/xml"

type confXmlModel struct {
	XMLName  xml.Name               `xml:"Service"`
	LocalSvc LocalServiceDefine     `xml:"LocalDefine"`
	SubSvcs  subSvcsModel           `xml:"SubSvcs"`
	SdDomain serviceDiscoveryDomain `xml:"ServiceDiscoveryDomain"`
}

type LocalServiceDefine struct {
	XMLName        xml.Name `xml:"LocalDefine"`
	Name           string   `xml:"Name"`
	NodeId         string   `xml:"NodeId"`
	LocalIP        string   `xml:"LocalIP"`
	Version        string   `xml:"Version"`
	HttpUrl        string   `xml:"HttpUrl"`
	WebPort        int      `xml:"WebPort"`
	UpdateInterval int      `xml:"UpdateInterval"`
}

type subSvcsModel struct {
	XMLName      xml.Name      `xml:"SubSvcs"`
	SubServices  []subsvcModel `xml:"Service"`
	SubsInterval int           `xml:"SubsInterval"`
}

type subsvcModel struct {
	XMLName xml.Name `xml:"Service"`
	Name    string   `xml:"Name"`
	Version string   `xml:"Version"`
}

type serviceDiscoveryDomain struct {
	XMLName xml.Name `xml:"ServiceDiscoveryDomain"`
	IP      string   `xml:"IP"`
	Port    int      `xml:"Port"`
	Timeout int      `xml:"Timeout"`
}

func (this *confXmlModel) fillWithXml(gson []byte) error {
	err := xml.Unmarshal(gson, this)
	if err != nil {
		return err
	}
	return nil
}
