package goldkey

import (
	"math/big"

	"github.com/xukgo/gsaber/encrypt/sm2"

	jsoniter "github.com/json-iterator/go"
)

type MachineUniqueInfo struct {
	MachineID        string `json:"machineId"`
	DiskSerialNumber string `json:"diskSn"` //硬盘序列号有可能为空
	CpuId            string `json:"cpuId"`
	Timestamp        int64  `json:"ts"`
}

func (this MachineUniqueInfo) ToJson() []byte {
	gson, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(this)
	return gson
}

func (this *MachineUniqueInfo) FillWithJson(data []byte) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(data, this)
}

func (this *MachineUniqueInfo) DecryptJson(data []byte, privd string) error {
	priv := new(sm2.PrivateKey)
	priv.Curve = sm2.GetSm2P256V1()
	priv.D, _ = new(big.Int).SetString(privd, 16)

	plainText, err := sm2.Decrypt(priv, data, sm2.C1C3C2)
	if err != nil {
		return err
	}
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(plainText, this)
}
