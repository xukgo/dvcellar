package perfScout

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

//返回应用使用的物理内存大小，单位kb
func getProcessMem(pid int) (int, error) {
	fileUrl := "/proc/" + strconv.Itoa(pid) + "/status"
	buf, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		fmt.Println("read /proc/pid/status fail", err)
	}

	sarr := strings.Split(string(buf), "\n")
	if len(sarr) < 22 {
		return -1, fmt.Errorf("/proc/pid/status file format error")
	}

	arr := strings.Split(sarr[21], " ")
	memStr := arr[len(arr)-2]
	m, err := strconv.Atoi(memStr)
	if err != nil {
		return -1, fmt.Errorf("/proc/pid/status file format error")
	}
	return m, nil
}
