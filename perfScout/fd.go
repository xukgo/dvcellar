package perfScout

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

//返回应用使用的fd句柄总数
func GetProcessFdCount(pid int) (int, error) {
	sum := 0
	dirUrl := "/proc/" + strconv.Itoa(pid) + "/fd/"
	rd, err := os.ReadDir(dirUrl)
	if err != nil {
		fmt.Println("read /proc/pid/fd/ fail", err)
		return -1, err
	}

	for _, fi := range rd {
		fiName := fi.Name()
		if strings.Index(fiName, ".") == 0 {
			continue
		}
		sum++
	}

	return sum, nil
}
