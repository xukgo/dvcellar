package perfScout

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func getTotalCpuTime() (int64,error) {
	buf, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		fmt.Println("read /proc/stat fail", err)
		return -1,err
	}

	sarr := strings.Split(string(buf), "\n")
	if len(sarr) == 0 {
		return -1,fmt.Errorf("/proc/stat file format error")
	}

	carr := strings.Split(sarr[0], " ")

	var sum int64 = 0
	for _, str := range carr {
		if str == "cpu" || len(str) == 0 {
			continue
		}

		c, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			continue
		}

		sum += c

	}
	//fmt.Printf("totoal cpu time:%v\r\n",sum)
	return sum,nil
}

func getProcCpuTime(pid int) (int64,error) {
	fileUrl := "/proc/" + strconv.Itoa(pid) + "/stat"
	buf, err := ioutil.ReadFile(fileUrl)
	if err != nil {
		fmt.Println("read /proc/pid/stat fail", err)
		return -1,err
	}

	carr := strings.Split(string(buf), " ")
	if len(carr) < 17{
		return -1,fmt.Errorf("/proc/pid/stat file format error")
	}

	var sum int64 = 0
	c, err := strconv.ParseInt(carr[13], 10, 64)
	if err == nil {
		sum += c
	}
	c, err = strconv.ParseInt(carr[14], 10, 64)
	if err == nil {
		sum += c
	}
	c, err = strconv.ParseInt(carr[15], 10, 64)
	if err == nil {
		sum += c
	}
	c, err = strconv.ParseInt(carr[16], 10, 64)
	if err == nil {
		sum += c
	}

	//fmt.Printf("totoal cpu time:%v\r\n",sum)
	return sum,nil
}
