package main

import (
	"TravelTimeEstimate/estimate"
	"strconv"
	"fmt"
	"time"
	"TravelTimeEstimate/defclass"
	"TravelTimeEstimate/preprocess"
)

// 测试Dijkstra算法
func testDijkstra(){
	pointSet := make(map[int64]*defclass.Point, 0)
	lineSet := make(map[int64]*defclass.Line, 0)
	for i := 1; i <= 8; i++ {
		p := defclass.NewPoint0()
		p.Id = int64(i)
		pointSet[p.Id] = p
	}
	lineSet[12] = defclass.NewLine5(pointSet[1], pointSet[2], 12, 12, 1)
	lineSet[13] = defclass.NewLine5(pointSet[1], pointSet[3], 13, 13, 1)
	lineSet[14] = defclass.NewLine5(pointSet[1], pointSet[4], 14, 14, 4)
	lineSet[16] = defclass.NewLine5(pointSet[1], pointSet[6], 16, 16, 2)
	lineSet[17] = defclass.NewLine5(pointSet[1], pointSet[7], 17, 17, 5)
	lineSet[26] = defclass.NewLine5(pointSet[2], pointSet[6], 26, 26, 2)
	lineSet[28] = defclass.NewLine5(pointSet[2], pointSet[8], 28, 28, 4)
	lineSet[37] = defclass.NewLine5(pointSet[3], pointSet[7], 37, 37, 3)
	lineSet[45] = defclass.NewLine5(pointSet[4], pointSet[5], 45, 45, 1)
	lineSet[56] = defclass.NewLine5(pointSet[5], pointSet[6], 56, 56, 1)
	lineSet[78] = defclass.NewLine5(pointSet[7], pointSet[8], 78, 78, 1)
	mymap := estimate.NewMapLoc2_2(pointSet, lineSet)
	dij := estimate.NewDijkstra(mymap)
	dij.Solve(4, 6)
	dij.PrintPathInfo(4, 6)
}

func testFindStreet(){
	findStreet := estimate.NewFindStreet()
	lineSet := findStreet.GetStreet()
	pointSet := findStreet.GetPointSet()
	for _, line := range lineSet {
		line.Print()
	}
	for _, point := range pointSet {
		point.Print()
	}
}

func testFindIntersection() {
	point := estimate.FindIntersection(40.7738924, -73.9716049)
	point.Print()
}

func testToSeconds(){
	tm := estimate.ToSeconds("2015-06-02 11:19:59.0")
	fmt.Println(tm)
}

func testGetHour(){
	hour := estimate.GetHour("2015-06-02 11:19:59.0")
	fmt.Println(hour)
}

func testAlgorithm_Estimate() {
	algorithm := estimate.NewAlgorithm(false)
	algorithm.Estimate(0, make([]int64, 5))
}

func testOpenStreetMap() {
	preprocess.ParseBigXML("E:\\TaxiQueryGenerator\\beijing_china.osm", "E:\\yellow_tripdata_2015-06.csv\\mongo\\MapPre")
}


// 估计行车时间
func travelTimeEstimate() {
	defer estimate.CloseSession()
	b := make([]int64, 5)
	var sum int64 = 0.0
	var sumRel float64 = 0.0 // 相对误差总和
	algorithm := estimate.NewAlgorithm(false)
	for i := 12; i < 24; i++ {
		tripi := "trip" + strconv.Itoa(i)
		algorithm.Preprocess(tripi)
		// 迭代开始时间
		fmt.Println("=========================================")
		startTime := time.Now().Unix()
		algorithm.Compute(tripi, make([]int64, 5))
		// 迭代耗时
		endTime := time.Now().Unix()
		fmt.Println("迭代耗时："+ strconv.FormatInt(endTime-startTime, 10) + "s")
		algorithm.Remain(i)
		sumRel += algorithm.Estimate(i, b)
		fmt.Println("=========================================")
	}
	for i := 0; i < 5; i++ {
		sum += b[i]
	}
	fmt.Println("=========================================")
	for i := 0; i < 5; i++ {
		fmt.Println(float64(b[i]) / float64(sum))
	}
	fmt.Println("平均相对误差为：" + fmt.Sprintf("%v", sumRel / float64(sum) * 100.0) + "%")
}
func main() {
	//trip.DropTrip("trip", 0, 23)
	//travelTimeEstimate()
	//trip.MdbTrip("E:\\15年5-6月黄车数据\\yellow_tripdata_2015-05.csv\\yellow_tripdata_2015-05.csv", "test")
	testOpenStreetMap()
}
