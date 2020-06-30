/**
 * @Author: zhangyw
 * @Description:
 * @File:  SubScribe
 * @Date: 2020/6/2 17:04
 */

package srvDiscover

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"log"
	"strings"
	"time"
)

type SubBasicInfo struct {
	Name      string
	Version   string
	Namespace string
}

func NewSubSrvBasicInfo(name, version, namespace string) *SubBasicInfo {
	model := new(SubBasicInfo)
	model.Name = name
	model.Version = version
	model.Namespace = namespace
	return model
}

type SrvNodeInfo struct {
	ModRevision   int64
	CacheUniqueId string
	RegInfo       RegisterInfo
}

type SubSrvNodeList struct {
	SubBasicInfo
	NodeInfos []*SrvNodeInfo
}

type SubscribeOption struct {
	Namespace string
}

var defaultSubscribeOption SubscribeOption = SubscribeOption{
	Namespace: "voice",
}

type SubscribeOptionFunc func(subscribeOp *SubscribeOption)

//func WithSubscribeNamespace(namespace string) SubscribeOptionFunc {
//	return func(subscribeOp *SubscribeOption) {
//		subscribeOp.Namespace = namespace
//	}
//}

/*
key格式： d9cloud.communication.SwitchServer/172.16.0.214/1602
先watch, 然后get全部一次，根据每个key的ModRevision进行更新
	watch的结果
 		PUT		只有ModRevision大于缓存里的，或者缓存里没有 才更新
		DELETE  只有ModRevision大于缓存里的， 才删除
	get全部结果
		get结果中，ModRevision大于缓存里的，更新
		get结果中，在缓存里没有的， 更新
		缓存有，get结果中没有的，删除
watch的更新和get全部的更新 需要加锁

watch的channel失败后，需要重新get全部一次
*/
func (this *Repo) SubScribe(subSrvInfos []SubBasicInfo, subcribeOptions ...SubscribeOptionFunc) error {
	serviceCount := len(subSrvInfos)
	if serviceCount <= 0 {
		return fmt.Errorf("subcribe names empty")
	}

	subcribeOp := new(SubscribeOption)
	*subcribeOp = defaultSubscribeOption

	for _, op := range subcribeOptions {
		op(subcribeOp)
	}

	for srvName, srvNodeList := range this.subsNodeCache {
		go this.watchSubs(srvName, srvNodeList, subcribeOp)
	}
	return nil
}

func (this *Repo) watchSubs(srvName string, srvNodeList *SubSrvNodeList, subscribeOp *SubscribeOption) {
	servicePrefix := fmt.Sprintf("registry.%s.%s", srvNodeList.Namespace, srvName)

	for {
		watchChan := this.client.Watch(clientv3.WithRequireLeader(context.TODO()), servicePrefix, clientv3.WithPrefix())
		if watchChan == nil {
			time.Sleep(time.Second)
			continue
		}

		//watch后必须进行一次成功的全查询
		err := this.getAll(srvName, srvNodeList)
		if err != nil {
			this.client.Watcher.Close()
			time.Sleep(time.Second)
			continue
		}

		//fmt.Println("watch begin ...")
		for watchResponse := range watchChan {
			this.updateByEvents(srvNodeList, watchResponse.Events)
		}
		//watchChan被关闭
		//fmt.Println("watchChan closed")
	}
}

func (this *Repo) getAll(srvName string, srvNodeList *SubSrvNodeList) error {
	servicePrefix := fmt.Sprintf("registry.%s.%s", srvNodeList.Namespace, srvName)
	getResponse, err := this.client.Get(context.TODO(), servicePrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	existKeyList := make([]string, 0, len(getResponse.Kvs))
	//更新插入
	for _, kv := range getResponse.Kvs {
		existKeyList = append(existKeyList, string(kv.Key))
		upsertNodeList(kv, srvNodeList)
	}

	//删除
	m := 0
	for idx := range srvNodeList.NodeInfos {
		if arrayKeyMatchUniqueId(existKeyList, srvNodeList.NodeInfos[idx].CacheUniqueId) >= 0 {
			srvNodeList.NodeInfos[m] = srvNodeList.NodeInfos[idx]
			m++
		}
	}
	srvNodeList.NodeInfos = srvNodeList.NodeInfos[:m]
	return nil
}

func (this *Repo) updateByEvents(srvNodeList *SubSrvNodeList, events []*clientv3.Event) {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, event := range events {
		switch event.Type {
		case mvccpb.PUT:
			//fmt.Println("put event ...")
			upsertNodeList(event.Kv, srvNodeList)
			break
		case mvccpb.DELETE:
			//fmt.Println("delete event ...")
			key := string(event.Kv.Key)
			modRevision := event.Kv.ModRevision
			srvNodeList.NodeInfos = removeNode(srvNodeList.NodeInfos, key, modRevision)
			break
		}
	}
}

func removeNode(infos []*SrvNodeInfo, key string, modRevision int64) []*SrvNodeInfo {
	m := 0
	for idx := range infos {
		s := infos[idx]
		if !checkKeyMatchNodeInfo(key, s.CacheUniqueId) || modRevision <= s.ModRevision {
			infos[m] = s
			m++
		}
	}
	return infos[:m]
}

func upsertNodeList(kv *mvccpb.KeyValue, srvNodeList *SubSrvNodeList) {
	key := string(kv.Key)
	valueBytes := kv.Value
	modRevision := kv.ModRevision

	for idx := range srvNodeList.NodeInfos {
		info := srvNodeList.NodeInfos[idx]
		if !checkKeyMatchNodeInfo(key, info.CacheUniqueId) {
			continue
		}

		if info.ModRevision >= modRevision {
			return
		}

		updateInfo := new(SrvNodeInfo)
		err := updateInfo.RegInfo.Deserialize(valueBytes)
		if err != nil {
			log.Printf("SrvNodeInfo unmarshal error:%s\r\n", err.Error())
			return
		}
		info.RegInfo = updateInfo.RegInfo
		info.ModRevision = modRevision
		return
	}

	newInfo := new(SrvNodeInfo)
	err := newInfo.RegInfo.Deserialize(valueBytes)
	if err != nil {
		log.Printf("SrvNodeInfo unmarshal error:%s\r\n", err.Error())
		return
	}
	newInfo.ModRevision = modRevision
	newInfo.CacheUniqueId = newInfo.RegInfo.UniqueId()
	srvNodeList.NodeInfos = append(srvNodeList.NodeInfos, newInfo)
}

func arrayKeyMatchUniqueId(array []string, id string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if checkKeyMatchNodeInfo(array[i], id) {
			index = i
			return
		}
	}
	return
}

func checkKeyMatchNodeInfo(key string, id string) bool {
	return strings.HasSuffix(key, id)
}

func (this *Repo) PrintAll() {
	this.locker.Lock()
	for srvName, srvNodeList := range this.subsNodeCache {
		fmt.Printf("-------------------- srv:%s len:%d\n", srvName, len(srvNodeList.NodeInfos))
		for _, node := range srvNodeList.NodeInfos {
			jsonBytes := node.RegInfo.Serialize()
			fmt.Println(node.ModRevision, node.CacheUniqueId, string(jsonBytes))
		}
	}

	this.locker.Unlock()
}

//func (this *ServiceDiscovery) Discover2(serviceEntry ...IServiceEntry) {
//	serviceCount := len(serviceEntry)
//	if serviceCount <= 0 {
//		return
//	}
//
//	this.AddDiscoverEntry(serviceEntry...)
//
//	for {
//		for _, service := range this.servcies {
//			this.getAll2(service)
//		}
//		time.Sleep(3 * time.Second)
//	}
//
//}

//func (this *ServiceDiscovery) getAll2(service *ServiceNodeList) {
//	servicePrefix := fmt.Sprintf("/registry/cpass/%s", service.Name)
//
//	contxt, _ := context.WithTimeout(context.TODO(), this.Timeout)
//	getResponse, err := this.client.Get(contxt, servicePrefix, clientv3.WithPrefix())
//	if err != nil {
//		return
//	}
//
//	serviceNodes := []*ServiceNode{}
//	for _, kv := range getResponse.Kvs {
//		srv := service.Alloctor.NewInstance()
//		err := srv.Deserialize(kv.Value)
//		if err != nil {
//			continue
//		}
//		serviceNodes = append(serviceNodes, &ServiceNode{
//			ModRevision: kv.ModRevision,
//			Entry:       srv,
//		})
//	}
//
//	service.Nodes = serviceNodes
//}
