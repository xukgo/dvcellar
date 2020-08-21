package srvDiscover

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	jsoniter "github.com/json-iterator/go"
	"github.com/xukgo/gsaber/encrypt/sm2"
	"log"
	"math/big"
	"time"
)

type Result struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (this Result) Clone() *Result {
	model := new(Result)
	model.Code = this.Code
	model.Description = this.Description
	return model
}

type LicResultInfo struct {
	Result          *Result `json:"result"`
	CallParallel    int     `json:"callParallel"`
	ExpireTimestamp int64   `json:"expire"` //单位秒
	Timestamp       int64   `json:"timestamp"`
}

func (this LicResultInfo) Clone() *LicResultInfo {
	model := new(LicResultInfo)
	if this.Result != nil {
		model.Result = this.Result.Clone()
	}
	model.CallParallel = this.CallParallel
	model.ExpireTimestamp = this.ExpireTimestamp
	model.Timestamp = this.Timestamp
	return model
}

type SubLicResultInfo struct {
	Info      *LicResultInfo
	Reversion int64
}

func (this *LicResultInfo) DecryptJson(data []byte, privKey string) error {
	priv := new(sm2.PrivateKey)
	priv.Curve = sm2.GetSm2P256V1()
	priv.D, _ = new(big.Int).SetString(privKey, 16)

	plainText, err := sm2.Decrypt(priv, data, sm2.C1C3C2)
	if err != nil {
		return err
	}
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(plainText, this)
}

const LIC_RESULT_KEY = "lic.result"

func (this *Repo) GetLicResult() *Result {
	this.licLocker.RLock()
	defer this.licLocker.RUnlock()

	if this.subLicResultInfo == nil {
		return nil
	}
	if this.subLicResultInfo.Info == nil {
		return nil
	}
	return this.subLicResultInfo.Info.Result.Clone()
}
func (this *Repo) GetLicResultInfo() *LicResultInfo {
	this.licLocker.RLock()
	defer this.licLocker.RUnlock()

	if this.subLicResultInfo == nil {
		return nil
	}
	if this.subLicResultInfo.Info == nil {
		return nil
	}
	return this.subLicResultInfo.Info.Clone()
}

func (this *Repo) StartSubLicResult(privKey string, watchFunc func(*LicResultInfo)) error {
	if len(privKey) == 0 {
		return fmt.Errorf("privKey is invalid")
	}

	this.licPrivkey = privKey
	this.licWatchFunc = watchFunc

	prefix := LIC_RESULT_KEY
	for {
		watchChan := this.client.Watch(clientv3.WithRequireLeader(context.TODO()), prefix, clientv3.WithPrefix())
		if watchChan == nil {
			time.Sleep(time.Second)
			continue
		}
		this.getLicResult()

		if this.licWatchFunc != nil {
			this.licWatchFunc(this.GetLicResultInfo())
		}

		for watchResponse := range watchChan {
			this.updateLicResultByEvents(watchResponse.Events)
			if this.licWatchFunc != nil {
				this.licWatchFunc(this.GetLicResultInfo())
			}
		}
	}
}

func (this *Repo) getLicResult() error {
	this.licLocker.Lock()
	defer this.licLocker.Unlock()

	prefix := LIC_RESULT_KEY
	getResponse, err := this.client.Get(context.TODO(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	//更新插入
	for _, kv := range getResponse.Kvs {
		this.upsertLicResult(kv)
	}
	return nil
}

func (this *Repo) updateLicResultByEvents(events []*clientv3.Event) {
	this.licLocker.Lock()
	defer this.licLocker.Unlock()

	for _, event := range events {
		switch event.Type {
		case mvccpb.PUT:
			this.upsertLicResult(event.Kv)
			break
		case mvccpb.DELETE:
			this.removeLicResult(event.Kv)
			break
		}
	}
}

func (this *Repo) removeLicResult(kv *mvccpb.KeyValue) {
	eventKey := string(kv.Key)
	if eventKey != LIC_RESULT_KEY {
		return
	}

	eventRevision := kv.ModRevision
	if this.subLicResultInfo == nil {
		this.subLicResultInfo = new(SubLicResultInfo)
		this.subLicResultInfo.Reversion = eventRevision
		return
	}
	if this.subLicResultInfo.Reversion <= eventRevision {
		this.subLicResultInfo.Info = nil
	}
}

func (this *Repo) upsertLicResult(kv *mvccpb.KeyValue) {
	eventKey := string(kv.Key)
	eventRevision := kv.ModRevision
	if eventKey != LIC_RESULT_KEY {
		return
	}

	if this.subLicResultInfo == nil {
		resultInfo, err := parseLicResult(kv.Value, this.licPrivkey)
		if err != nil {
			log.Println("parse lic result error:", err.Error())
		} else {
			this.subLicResultInfo = new(SubLicResultInfo)
			this.subLicResultInfo.Reversion = eventRevision
			this.subLicResultInfo.Info = resultInfo
		}
		return
	}

	if this.subLicResultInfo.Reversion >= eventRevision {
		return
	}

	resultInfo, err := parseLicResult(kv.Value, this.licPrivkey)
	if err != nil {
		log.Println("parse lic result error:", err.Error())
	} else {
		this.subLicResultInfo.Reversion = eventRevision
		this.subLicResultInfo.Info = resultInfo
	}
}

func parseLicResult(data []byte, privKey string) (*LicResultInfo, error) {
	data, err := hex.DecodeString(string(data))
	if err != nil {
		log.Println("sub licResult value decode hexString error:", err.Error())
		return nil, err
	}

	model := new(LicResultInfo)
	err = model.DecryptJson(data, privKey)
	if err != nil {
		log.Println("sub value licResult DecryptJson  error:", err.Error())
		return nil, err
	}

	if model.Result == nil {
		log.Println("sub value licResult unmarshal json result  error")
		return model, nil
	}
	if model.Result.Code != 0 {
		return model, nil
	}

	//有效的还要校验下时间戳
	//15分钟没有更新，则认定许可有问题
	dtNow := time.Now()
	secSub := dtNow.Sub(time.Unix(model.Timestamp, 0)).Seconds()
	if secSub > 15*60 {
		model.Result.Code = 3001
		model.Result.Description = "update timestamp expired"
	}

	secSub = dtNow.Sub(time.Unix(model.ExpireTimestamp, 0)).Seconds()
	if secSub > 15*60 {
		model.Result.Code = 3002
		model.Result.Description = "lic timestamp expired"
	}
	return model, nil
}
