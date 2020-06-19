/**
 * @Author: zhangyw
 * @Description:
 * @File:  Register
 * @Date: 2020/6/2 17:04
 */

package srvDiscover

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func (this *Repo) Register(srvInfo *RegisterInfo, options ...RegisterOptionFunc) {
	regOption := new(RegisterOption)
	*regOption = defaultRegisterOption

	for _, op := range options {
		op(regOption)
	}

	var lease *clientv3.LeaseGrantResponse = nil
	var err error

	for {
		ctx, _ := context.WithTimeout(context.TODO(), regOption.ConnTimeout)
		lease, err = this.client.Grant(ctx, regOption.TTLSec)
		if err != nil || lease == nil {
			time.Sleep(time.Second)
			continue
		}

		this.fillRegMoudleInfo(srvInfo, regOption.BeforeRegister)
		err := this.clientUpdateLeaseContent(lease, srvInfo, regOption)
		if err != nil {
			this.client.Lease.Close()
			time.Sleep(time.Second)
			continue
		}

		this.KeepaliveLease(lease, srvInfo, regOption)
	}
}

func (this *Repo) KeepaliveLease(lease *clientv3.LeaseGrantResponse, srvInfo *RegisterInfo, regOption *RegisterOption) {
	keepaliveChan, err := this.client.KeepAlive(context.TODO(), lease.ID) //这里需要一直不断，context不允许设置超时
	if err != nil || keepaliveChan == nil {
		time.Sleep(time.Second)
		return
	}

	timeSaved := time.Now()
	for {
		select {
		case keepaliveResponse, ok := <-keepaliveChan:
			if !ok || keepaliveResponse == nil {
				fmt.Println(">>>error keepaliveResponse")
				return
			}
			//fmt.Println("keepaliveResponse", keepaliveResponse)
			break
		default:
			if !regOption.AlwaysUpdate {
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			if time.Since(timeSaved) < regOption.Interval {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			this.fillRegMoudleInfo(srvInfo, regOption.BeforeRegister)
			err := this.clientUpdateLeaseContent(lease, srvInfo, regOption)
			if err != nil {
				this.client.Lease.Close()
				return
			}

			timeSaved = time.Now()
		}
	}
}

func (this *Repo) clientUpdateLeaseContent(lease *clientv3.LeaseGrantResponse, srvInfo *RegisterInfo, regOption *RegisterOption) interface{} {
	key := srvInfo.FormatRegisterKey(regOption.Namespace)
	value := srvInfo.Serialize()
	valueStr := string(value)

	//fmt.Println("keep", key, valueStr)
	_, err := this.client.Put(context.TODO(), key, valueStr, clientv3.WithLease(lease.ID))
	return err
}

func (this *Repo) fillRegMoudleInfo(info *RegisterInfo, beforeRegisterFunc BeforeRegisterFunc) {
	if beforeRegisterFunc != nil {
		beforeRegisterFunc(info)
	}
	info.Global.State = "online"
	info.Global.RefreshTimestamp(time.Now())
}
