package goldkey

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type MountDirInfo struct {
	Dev string
	Dir string
}

func GetDiskDevName(mountDir string) string {
	mounts := GetDiskDevMountDirs()
	selectDevName := ""
	for idx := range mounts {
		if mounts[idx].Dir == mountDir {
			selectDevName = mounts[idx].Dev
			break
		}
	}
	return selectDevName
}
func GetDiskDevMountDirs() []MountDirInfo {
	content, err := ioutil.ReadFile("/etc/mtab")
	if err != nil {
		fmt.Println("read file error:", err.Error())
		return nil
	}
	arr := strings.Split(string(content), "\n")
	infos := make([]MountDirInfo, 0, 4)
	for _, s := range arr {
		arr1 := strings.Split(s, " ")
		if len(arr1) < 2 {
			continue
		}
		dev := arr1[0]
		if !strings.HasPrefix(dev, "/dev/") {
			continue
		}
		model := MountDirInfo{
			Dev: dev,
			Dir: arr1[1],
		}
		infos = append(infos, model)
	}
	return infos
}
