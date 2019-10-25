package svcDiscovery

type Observer interface {
	UpdateFromSvcDisc(*ServieDiscoveryRepo)
}
