package goldkey

import (
	"fmt"
	"math/big"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/xukgo/gsaber/encrypt/sm2"
)

type BasicOfflineLicense struct {
	MachineInfo     *MachineUniqueInfo `json:"machine"`
	LimitSvcCount   int                `json:"limitSvcCount"`
	SignTimestamp   int64              `json:"signTs"`
	ExpireTimestamp int64              `json:"expireTs"`
}

func (this BasicOfflineLicense) ToPrettyJson() []byte {
	gson, _ := jsoniter.ConfigCompatibleWithStandardLibrary.MarshalIndent(this, "", "   ")
	return gson
}
func (this BasicOfflineLicense) ToJson() []byte {
	gson, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(this)
	return gson
}

func (this *BasicOfflineLicense) EncryptJson(pubx, puby string) ([]byte, error) {
	gson := this.ToJson()
	pub := new(sm2.PublicKey)
	pub.Curve = sm2.GetSm2P256V1()
	pub.X, _ = new(big.Int).SetString(pubx, 16)
	pub.Y, _ = new(big.Int).SetString(puby, 16)

	cipherText, err := sm2.Encrypt(pub, gson, sm2.C1C3C2)
	return cipherText, err
}

func (this *BasicOfflineLicense) DecryptJson(data []byte, privd string) error {
	priv := new(sm2.PrivateKey)
	priv.Curve = sm2.GetSm2P256V1()
	priv.D, _ = new(big.Int).SetString(privd, 16)

	plainText, err := sm2.Decrypt(priv, data, sm2.C1C3C2)
	if err != nil {
		return err
	}
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(plainText, this)
}

func (this *BasicOfflineLicense) Print() {
	dt := time.Unix(0, this.MachineInfo.Timestamp)
	fmt.Printf("许可节点，唯一码=%s，硬盘序列号=%s，cpuID=%s，生成时间=%s\n",
		this.MachineInfo.MachineID, this.MachineInfo.DiskSerialNumber, this.MachineInfo.CpuId, dt.Format("2006-01-02 15:04:05"))
	fmt.Printf("限制节点数量:%d\n", this.LimitSvcCount)
	dt = time.Unix(this.SignTimestamp, 0)
	fmt.Printf("许可签发时间:%s\n", dt.Format("2006-01-02 15:04:05"))
	fmt.Printf("过期时间戳:%d\n", this.ExpireTimestamp)
	dt = time.Unix(this.ExpireTimestamp, 0)
	fmt.Printf("过期时间预估:%s\n", dt.Format("2006-01-02 15:04:05"))

	subTime := dt.Sub(time.Now())
	hourVal := subTime.Hours()
	dayCount := int(hourVal) / 24
	hourCount := hourVal - float64(dayCount)*24
	fmt.Printf("有效服务时间:%d天%.2f小时\n", dayCount, hourCount)
	if time.Now().Unix() > this.ExpireTimestamp {
		fmt.Printf("！！！警告:配置的服务时间已过期\n")
	}
}
