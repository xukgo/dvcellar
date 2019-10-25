package perfScout

import (
	"io/ioutil"
	"strconv"
	"strings"
)

//返回应用使用的fd句柄总数
func getProcessFdCount(pid int) int {
	sum := 0
	dirUrl := "/proc/" + strconv.Itoa(pid) + "/fd/"
	rd, err := ioutil.ReadDir(dirUrl)
	if err != nil {
		return -1
	}

	for _, fi := range rd {
		fiName := fi.Name()
		if strings.Index(fiName, ".") == 0 {
			continue
		}
		sum++
	}

	return sum
}
