package preprocess

import (
	"TravelTimeEstimate/estimate"
	"os"
	"fmt"
	"bufio"
	"io"
	"strings"
	"strconv"
	"TravelTimeEstimate/defclass"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
)
// 将MapPre处理好的数据存入数据库
func MdbInsert(mdbPath string) {
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.MAP_DB)
	collArc := db.C("mapArc")
	collPoint := db.C("mapPoint")
	fpoint, err := os.Open(mdbPath + "\\mapPoint.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer fpoint.Close()
	farc, err := os.Open(mdbPath + "\\mapArc.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer farc.Close()
	pEdge := make(map[int64][]int64)
	// 向数据库插入边的信息
	buf := bufio.NewReader(farc)
	for {
		value, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		stk := strings.Fields(value)
		aid, _ := strconv.ParseInt(stk[0], 10, 64)
		x, _ := strconv.ParseInt(stk[1], 10, 64)
		y, _ := strconv.ParseInt(stk[2], 10, 64)
		length, _ := strconv.ParseFloat(stk[3], 64)
		wayid, _ := strconv.ParseInt(stk[4], 10, 64)
		if _, ok := pEdge[x]; ok {
			pEdge[x] = append(pEdge[x], aid)
		} else {
			slice := make([]int64, 0)
			slice = append(slice, aid)
			pEdge[x] = slice
		}
		if _, ok := pEdge[y]; ok {
			pEdge[y] = append(pEdge[y], aid)
		} else {
			slice := make([]int64, 0)
			slice = append(slice, aid)
			pEdge[y] = slice
		}
		line := defclass.NewLine0()
		line.Sid = aid
		line.Gis = bson.M{"x": x, "y": y}
		line.Length = length
		line.Wayid = wayid
		collArc.Insert(line)
	}

	// 向数据库插入点的信息
	buf = bufio.NewReader(fpoint)
	for {
		value, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		stk := strings.Fields(value)
		pid, _ := strconv.ParseInt(stk[0], 10, 64)
		lat, _ := strconv.ParseFloat(stk[1], 64)
		lon, _ := strconv.ParseFloat(stk[2], 64)
		point := defclass.NewPoint0()
		point.Id = pid
		point.Gis = bson.D{{"lon", lon}, {"lat", lat}} // 要采用有顺序的，方便建立索引
		point.Line_set = pEdge[pid]
		collPoint.Insert(point)
	}
	     index := mgo.Index{
	         Key: []string{"$2dsphere:gis"},
	         Bits: 26,
	     }
	collPoint.EnsureIndex(index)
	fmt.Println("插入数据库完成！")
}
