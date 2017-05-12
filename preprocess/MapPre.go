package preprocess

import (
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	"strconv"
	"list"
	"regexp"
	"time"
	"TravelTimeEstimate/estimate"
)
//设定一个UUID 用于随机生成ID （随机ID为当前时间精确到毫秒+UUID） UUID为递增的
var uuid int64 = 10000
var modnum int64 = uuid * 10

// 对抽取的地图数据作额外处理，以便存入mongodb数据库
func MapPre(inputFilePath string, outputFilePath string) {
	finArc, err := os.Open(inputFilePath + "\\Arc.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer finArc.Close()
	finPoint, err := os.Open(inputFilePath + "\\Point.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer finPoint.Close()
	foutPoint, err := os.Create(outputFilePath + "\\mapPoint.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer foutPoint.Close()
	foutArc, err := os.Create(outputFilePath + "\\mapArc.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer foutArc.Close()
	pidMap := make(map[string]int64, 0)
	pidMapOrder := make([]string, 0) // 用于记录存入pidMap中的数据的顺序
	point := make(map[string][]string, 0)
	var preValue int64 = -1
	var lat1 float64 = -1
	var lon1 float64 = -1
	var preKey string

	highList := list.NewArrayList()
	// 找出highway的ID添加到highList中
	buf := bufio.NewReader(finArc)
	for {
		value, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		stk := strings.Fields(value)
		highWay, _ := strconv.ParseInt(stk[0], 10, 64)
		if "highway" == stk[2] {
			highList.Add(highWay)
		}
	}
	// 找出highway对应点的ID
	finArc.Seek(0,0)
	buf = bufio.NewReader(finArc)
	for {
		value, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		var pointID string
		stk := strings.Fields(value)
		highWay, _ := strconv.ParseInt(stk[0], 10, 64)
		if highList.Contains(highWay) {
			if match, _ := regexp.MatchString("^\\d+$", stk[2]); match {	//获取点道路上的点ID
				pointID = stk[2]
				if _, ok := pidMap[pointID]; ok {
					flag := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
					uuid = (uuid + 1) % modnum
					pointID += " " + flag + strconv.FormatInt(uuid, 10)
					pidMap[pointID] = highWay
					pidMapOrder = append(pidMapOrder, pointID)
				} else {
					pidMap[pointID] = highWay
					pidMapOrder = append(pidMapOrder, pointID)
				}
			}
		}
	}
	fmt.Println("查找highway完毕！")
	// 找出点的信息
	buf = bufio.NewReader(finPoint)
	out := bufio.NewWriter(foutPoint)
	for {
		value, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		stk := strings.Fields(value)
		pointID := stk[0]
		_, ok1 := pidMap[pointID]
		_, ok2 := point[pointID]
		if ok1 && !ok2 {
			lat := stk[1]
			lon := stk[2]
			point[pointID] = []string{lat, lon}
			str := pointID + "	" + lat + "	" + lon + "\n"
			out.WriteString(str)
			out.Flush()
		}
	}
	fmt.Println("输出点信息完毕！")
	// 找出线段的信息
	out = bufio.NewWriter(foutArc)
	for _, v := range pidMapOrder {
		strKey := strings.Fields(v)
		key := strKey[0]
		slice, ok := point[key]
		if !ok || len(slice) < 2 {
			continue
		}
		lat2, _ := strconv.ParseFloat(slice[0], 64)
		lon2, _ := strconv.ParseFloat(slice[1], 64)
		if pidMap[v] == preValue {
			length := estimate.Distance(lat1, lon1, lat2, lon2)
			flag := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
			uuid = (uuid + 1) % modnum
			arcId := flag + strconv.FormatInt(uuid, 10)
			str := arcId + " " + preKey + " " + key + " " + fmt.Sprintf("%v", length) + " " + strconv.FormatInt(pidMap[v], 10) + "\n"
			out.WriteString(str)
			out.Flush()
		} else {
			preValue = pidMap[v]
		}
		preKey = key
		lat1 = lat2
		lon1 = lon2
	}
	fmt.Println("输出弧段信息完毕！")
}
