package estimate

import (
	"TravelTimeEstimate/defclass"
	"set"
	"list"
	"strconv"
	"math"
	"fmt"
	"gopkg.in/mgo.v2/bson"
)
/**
 * 估计行车时间的主要算法类（不同边采用不同的k值调整）多次迭代调整    添加限制，保证最短路径跟真实行驶距离相差不大
 * @author Administrator
 *
 */
type Algorithm struct {
	lineSet map[int64]*defclass.Line	// 道路集合
	pointSet map[int64]*defclass.Point	// 点集合
	tripStreet set.Set			// 旅途经过的所有道路的id
	tripList list.List			// 存放所有trip信息
	dij *Dijkstra
}

func NewAlgorithm(needMergeEquivalentTrip bool) *Algorithm {
	if needMergeEquivalentTrip {
		// 过滤trip，并将过滤后的trip存入数据库，只需执行一次
		MergeEquivalentTrip(USE_WEEK)	// 按时间分片并过滤后的Trip数据
	}
	return &Algorithm{}
}

// 初始化路径计算，同时对trip进行第二次的过滤
func (algorithm *Algorithm) Preprocess(tripi string) {
	if algorithm.dij == nil {
		findStreet := NewFindStreet("mapIntersection")
		algorithm.lineSet = findStreet.GetStreet()
		algorithm.pointSet = findStreet.GetPointSet()
		mymap := NewMapLoc2_2(algorithm.pointSet, algorithm.lineSet)
		algorithm.dij = NewDijkstra(mymap)
	}
	algorithm.tripList = list.NewArrayList()
	session := GetSesson()
	defer session.Close()
	db1 := session.DB(TRIP_DB)
	dbcoll := db1.C(tripi)
	iter := dbcoll.Find(nil).Iter()
	//查询是否计算过
	db2 := session.DB(MAP_DB)
	coll := db2.C("shortestRoute")
	result := defclass.NewSelectTrip0()
	for iter.Next(result) {
		result.Distance = result.Distance * 1600.0
		cursor := coll.FindId(strconv.FormatInt(result.Sid, 10) + "-" + strconv.FormatInt(result.Eid, 10))
		if cnt, _ := cursor.Count(); cnt > 0 {	// 如果已经算过
			route := defclass.NewShortestRoute()
			cursor.One(route)
			averageSpeed := route.Length / result.TravelTime // 计算平均速度
			if averageSpeed < 0.5 || averageSpeed > 30 || math.Abs(route.Length - result.Distance) > result.Distance * ERROR { // 移除速度太慢或太快的旅途
				continue
			}
			trip := defclass.NewTrip0()
			trip.SetTravelTime(result.TravelTime)
			trip.Length = route.Length // 最短路径长度
			trip.SetPath(route.Path)
			algorithm.tripList.Add(trip)
			continue
		}
		cursor = coll.FindId(strconv.FormatInt(result.Eid, 10) + "-" + strconv.FormatInt(result.Sid, 10))
		if cnt, _ := cursor.Count(); cnt > 0 {	// 如果已经算过
			route := defclass.NewShortestRoute()
			cursor.One(route)
			averageSpeed := route.Length / result.TravelTime // 计算平均速度
			if averageSpeed < 0.5 || averageSpeed > 30 || math.Abs(route.Length - result.Distance) > result.Distance * ERROR { // 移除速度太慢或太快的旅途
				continue
			}
			n := len(route.Path)
			res := make([]interface{}, n)
			for i := n-1; i >= 0; i-- {	// 反转
				res[n-1-i] = route.Path[i]
			}
			dbroute := defclass.NewShortestRoute()
			dbroute.Id = strconv.FormatInt(result.Sid, 10) + "-" + strconv.FormatInt(result.Eid, 10)
			dbroute.Path = res
			dbroute.Length = route.Length
			coll.Insert(dbroute)
			trip := defclass.NewTrip0()
			trip.SetTravelTime(result.TravelTime)
			trip.Length = route.Length // 最短路径长度
			trip.SetPath(res)
			algorithm.tripList.Add(trip)
			continue
		}
		length, path := algorithm.dij.Solve(result.Sid, result.Eid)
		averageSpeed := length / result.TravelTime // 计算平均速度
		if averageSpeed < 0.5 || averageSpeed > 30 || math.Abs(length - result.Distance) > result.Distance * ERROR { // 移除速度太慢或太快的旅途
			continue
		}
		trip := defclass.NewTrip0()
		trip.SetTravelTime(result.TravelTime)
		trip.Length = length
		trip.SetPath(path)
		algorithm.tripList.Add(trip)
	}
}

/**
 * 算法的计算迭代，返回计算的相对误差总和
 * @param tripi 要计算的时间片
 */
func (algorithm *Algorithm) Compute(tripi string, b []int64) float64 {
	algorithm.tripStreet = set.NewHashSet()
	n := algorithm.tripList.Len()
	fmt.Println("Trip的个数为：" + strconv.Itoa(n))
	ts := make(map[int64]list.List, 0)	// 用于存放道路和所有经过它的Trip的集合的映射关系
	var relErr float64 = 0.0
	for i := 0; i < n; i++ {
		trip := algorithm.tripList.Get(i).(*defclass.Trip)
		path := trip.GetPath()
		var et float64 = 0.0	// 估计的旅途行车时间
		for _, v := range path {
			id := v.(int64)
			et += algorithm.lineSet[id].GetTravelTime()
			if _, ok := ts[id]; ok {
				ts[id].Add(trip)
			} else {
				list := list.NewArrayList()
				list.Add(trip)
				ts[id] = list
			}
		}
		trip.Et = et
		relErr += math.Abs(trip.Et - trip.GetTravelTime()) / trip.GetTravelTime()
	}
	for k, _ := range ts {
		algorithm.tripStreet.Add(k)
	}
	var initK float64 = 2.0
	mapk := make(map[int64]float64, 0)
	mapos := make(map[int64]float64, 0)
	for true {
		flag := true
		for id, trips := range ts {
			if _, ok := mapk[id]; !ok {
				mapk[id] = initK
			}
			if _, ok := mapos[id]; mapk[id] < 1.0001 || (ok && mapos[id] == 0) {
				continue
			} else {
				flag = false
			}
			var os float64 = 0.0
			size := trips.Len()
			for i := 0; i < size; i++ {
				trip := trips.Get(i).(*defclass.Trip)
				os += (trip.Et - trip.GetTravelTime()) / trip.GetTravelTime()
			}
			mapos[id] = os
			if os < 0 {
				algorithm.lineSet[id].SetTravelTime(algorithm.lineSet[id].GetTravelTime() * mapk[id])
			} else if os > 0 {

				algorithm.lineSet[id].SetTravelTime(algorithm.lineSet[id].GetTravelTime() / mapk[id])
			}
		}
		var newRelErr float64 = 0.0 // 新的相对误差
		newetSlice := make([]float64, n)	// 用于存放新估计的旅途行车时间，要和trip保持同样的顺序
		for i := 0; i < n; i++ {
			trip := algorithm.tripList.Get(i).(*defclass.Trip)
			path := trip.GetPath()
			var newet float64 = 0.0 // 新估计的旅途行车时间
			for _, v := range path {
				id := v.(int64)
				newet += algorithm.lineSet[id].GetTravelTime()
			}
			newetSlice[i] = newet
			newRelErr += math.Abs(newet - trip.GetTravelTime()) / trip.GetTravelTime()
		}
		if newRelErr < relErr {	// 新的估计比之前的好
			for i := 0; i < n; i++ {
				trip := algorithm.tripList.Get(i).(*defclass.Trip)
				trip.Et = newetSlice[i]	// 更新旅途行车时间的估计
			}
			for id, _ := range ts {
				mapk[id] = initK
			}
			relErr = newRelErr
		} else {	// 新的估计比之前的坏
			for id, _ := range ts {
				if mapos[id] < 0 {
					algorithm.lineSet[id].SetTravelTime(algorithm.lineSet[id].GetTravelTime() / mapk[id])
				} else if mapos[id] > 0 {
					algorithm.lineSet[id].SetTravelTime(algorithm.lineSet[id].GetTravelTime() * mapk[id])
				} else {
					continue
				}
				mapk[id] = 1 + (mapk[id] - 1) * 0.5 // 减少道路行车时间的增加/减少步长
			}
		}
		if flag {	// k太小，跳出 内部循环
			break
		}
	}
	var sumRel float64 = 0.0
	for i := 0; i < n; i++ {
		trip := algorithm.tripList.Get(i).(*defclass.Trip)
		error := math.Abs(trip.Et - trip.GetTravelTime()) / trip.GetTravelTime() * 100.0
		sumRel += error / 100.0
		if error < 10 {
			b[0]++
		} else if error < 20 {
			b[1]++
		} else if error < 30 {
			b[2]++
		} else if error < 40 {
			b[3]++
		} else {
			b[4]++
		}
	}
	fmt.Println(tripi + "平均相对误差为：" + fmt.Sprintf("%v", sumRel / float64(n) * 100.0) + "%")
	return sumRel
}

/**
 * 计算余下道路的行车时间
 * @param i 第几个时间段
 */
func (algorithm *Algorithm) Remain(i int) {
	nTripStreet := make([]int64, 0)	// 旅途没有经过的其他道路
	for id, _ := range algorithm.lineSet {
		if !algorithm.tripStreet.Contains(id) {
			nTripStreet = append(nTripStreet, id)
		}
	}
	// 第一次计算
	nmap := make(map[int64][]int64, 0)
	for _, id := range nTripStreet {
		sid := algorithm.lineSet[id].GetP()[0].Id
		eid := algorithm.lineSet[id].GetP()[1].Id
		slice := make([]int64, 0)
		ids := algorithm.tripStreet.Elements()
		for _, v := range ids {
			streetId := v.(int64)
			sid1 := algorithm.lineSet[streetId].GetP()[0].Id
			eid1 := algorithm.lineSet[streetId].GetP()[1].Id
			if sid1==sid || sid1==eid || eid1==sid || eid1==eid {
				slice = append(slice, streetId)
			}
		}
		nmap[id] = slice
	}
	for len(nmap) > 0 {
		// 寻找切片大小最大的
		var maxid int64
		var maxSize int = 0
		for k, _ := range nmap {
			if len(nmap[k]) > maxSize {
				maxSize = len(nmap[k])
				maxid = k
			}
		}
		if len(nmap[maxid]) == 0 {
			break
		}
		var v float64 = 0.0	// 道路速度
		for _, id := range nmap[maxid] {
			if algorithm.lineSet[id].Length == 0.0 && algorithm.lineSet[id].GetTravelTime() == 0.0 {	// 避免NaN的出现
				continue
			}
			v += algorithm.lineSet[id].Length / algorithm.lineSet[id].GetTravelTime()
		}
		if v == 0.0 {	// 不再计算保留原来的值
			delete(nmap, maxid)
			continue
		}
		v /= float64(len(nmap[maxid]))
		algorithm.lineSet[maxid].SetTravelTime(algorithm.lineSet[maxid].Length / v)
		delete(nmap, maxid)	// 移除已经计算的
		// 更新nmap中已经计算过切片中的道路id
		sid := algorithm.lineSet[maxid].GetP()[0].Id
		eid := algorithm.lineSet[maxid].GetP()[1].Id
		for id, _ := range nmap {
			sid1 := algorithm.lineSet[id].GetP()[0].Id
			eid1 := algorithm.lineSet[id].GetP()[1].Id
			if sid1 == sid || sid1 == eid || eid1 == sid || eid1 == eid{
				nmap[id] = append(nmap[id], maxid)
			}
		}
	}
}

func (algorithm *Algorithm) Estimate(i int, b []int64) float64 {
	var sumRel float64 = 0.0
	prefixs := []string{"00","01","02","03","04","05","06","07","08","09","10","11","12",
		"13","14","15","16","17","18","19","20","21","22","23"}
	// 用结果估计所有trip
	session := GetSesson()
	defer session.Close()
	db := session.DB(TRIP_DB)
	dbcoll := db.C(ES_WEEK_DAY)
	re1 := "(\\d)"	// Any Single Digit 1
	re2 := "(\\d)"	// Any Single Digit 2
	re3 := "(\\d)"	// Any Single Digit 3
	re4 := "(\\d)"	// Any Single Digit 4
	re5 := "(-)"	// Any Single Character 1
	re6 := "(\\d)"	// Any Single Digit 5
	re7 := "(\\d)"	// Any Single Digit 6
	re8 := "(-)"	// Any Single Character 2
	re9 := "(\\d)"	// Any Single Digit 7
	re10 := "(\\d)"	// Any Single Digit 8
	re11 := "(\\s+)"	// White Space 1
	prefix := re1+re2+re3+re4+re5+re6+re7+re8+re9+re10+re11+prefixs[i]
	cursor := dbcoll.Find(bson.M{"pickuptime": bson.M{"$regex":"^"+prefix+":.*$"}})
	iter := cursor.Iter()
	var cnt int64 = 0
	tripData := defclass.NewTripData()
	for iter.Next(tripData) {
		tripData.Distance = tripData.Distance * 1600.0
		pickup_lat := tripData.Pickup["lat"].(float64)
		pickup_lon := tripData.Pickup["lon"].(float64)
		dropoff_lat := tripData.Dropoff["lat"].(float64)
		dropoff_lon := tripData.Dropoff["lon"].(float64)
		o := FindIntersection(pickup_lat, pickup_lon)
		if o == nil {
			continue
		}
		d := FindIntersection(dropoff_lat, dropoff_lon)
		if d == nil {
			continue
		}
		if o.Id == d.Id {	// 如果是回路,去除回路
			continue
		}
		time := float64(ToSeconds(tripData.Dorpofftime) - ToSeconds(tripData.Pickuptime))
		if time < 120 {
			continue
		}
		path := algorithm.getPath(o.Id, d.Id)
		var et float64 = 0.0 // 估计的旅途行车时间
		var length float64 = 0.0 // 路径长度
		for _, v := range path {
			id := v.(int64)
			et += algorithm.lineSet[id].GetTravelTime()
			length += algorithm.lineSet[id].Length
		}
		if math.Abs(length - tripData.Distance) > tripData.Distance * ERROR {
			continue
		}
		error := math.Abs(et - time) / time * 100.0
		sumRel += error / 100.0
		cnt++
		if error < 10 {
			b[0]++
		} else if error < 20 {
			b[1]++
		} else if error < 30 {
			b[2]++
		} else if error < 40 {
			b[3]++
		} else {
			b[4]++
		}
	}
	fmt.Println("时间段" + strconv.Itoa(i) + "平均相对误差为：" + fmt.Sprintf("%v", sumRel / float64(cnt) * 100.0) + "%")
	return sumRel
}

func (algorithm *Algorithm) getPath(sid int64, eid int64) []interface{} {
	session := GetSesson()
	defer session.Close()
	db := session.DB(MAP_DB)
	coll := db.C("shortestRoute")
	//查询是否计算过
	cursor := coll.FindId(strconv.FormatInt(sid, 10) + "-" + strconv.FormatInt(eid, 10))
	if cnt, _ := cursor.Count(); cnt > 0 {
		route := defclass.NewShortestRoute()
		cursor.One(route)
		return route.Path
	}
	cursor = coll.FindId(strconv.FormatInt(eid, 10) + "-" + strconv.FormatInt(sid, 10))
	if cnt, _ := cursor.Count(); cnt > 0 {
		route := defclass.NewShortestRoute()
		cursor.One(route)
		return route.Path
	}
	_, path := algorithm.dij.Solve(sid, eid)
	return path
}