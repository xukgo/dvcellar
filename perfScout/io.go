package perfScout

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

/*cat /proc/$PID/io
rchar:  读出的总字节数，read或者pread（）中的长度参数总和（pagecache中统计而来，不代表实际磁盘的读入）
wchar: 写入的总字节数，write或者pwrite中的长度参数总和
syscr:  read（）或者pread（）总的调用次数
syscw: write（）或者pwrite（）总的调用次数
read_bytes: 实际从磁盘中读取的字节总数   （这里if=/dev/zero 所以没有实际的读入字节数）
write_bytes: 实际写入到磁盘中的字节总数
cancelled_write_bytes: 由于截断pagecache导致应该发生而没有发生的写入字节数（可能为负数）
*/

type iorwInfo struct {
	nanoStamp  int64
	readBytes  int64
	writeBytes int64
}

//返回应用进行的硬盘io操作的读写总历史字节数,返回
func GetIOBytes(pid int) (iorwInfo, error) {
	resInfo := iorwInfo{
		nanoStamp: time.Now().UnixNano(),
	}

	fileUrl := "/proc/" + strconv.Itoa(pid) + "/io"
	buf, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		fmt.Println("read /proc/pid/io fail", err)
		return resInfo, err
	}

	sarr := strings.Split(string(buf), "\n")
	if len(sarr) < 6 {
		return resInfo, fmt.Errorf("/proc/pid/io file format error")
	}

	arr := strings.Split(sarr[4], " ")
	resInfo.readBytes, err = strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		return resInfo, err
	}

	arr = strings.Split(sarr[5], " ")
	resInfo.writeBytes, err = strconv.ParseInt(arr[1], 10, 64)
	if err != nil {
		return resInfo, err
	}
	return resInfo, nil
}
