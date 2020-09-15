package goldkey

import (
	"fmt"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
)

func Info() *MachineUniqueInfo {
	info := new(MachineUniqueInfo)
	id, err := machineid.ID()
	if err != nil {
		//fmt.Println("获取机器唯一码失败：", err)
		id = ""
	}
	info.MachineID = id

	devName := GetDiskDevName("/")
	diskSn := disk.GetDiskSerialNumber(devName)
	info.DiskSerialNumber = diskSn

	cpuInfos, err := cpu.Info()
	if err != nil {
		fmt.Println("获取cpu信息失败：", err)
		return nil
	}
	info.CpuId = fmt.Sprintf("%s_%s_%d_%s", cpuInfos[0].Family, cpuInfos[0].Model, cpuInfos[0].Stepping, cpuInfos[0].Microcode)
	info.Timestamp = time.Now().UnixNano()
	return info
}
