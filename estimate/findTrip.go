package estimate

import (
	"math"
	"TravelTimeEstimate/defclass"
	"gopkg.in/mgo.v2/bson"
	"time"
	"list"
	"strconv"
)
/*
按时间段找出符合要求的Trip
 */
func rad(d float64) float64 {
	return d * PI / 180.0
}

// 计算两点间的距离
func Distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radLat1 := rad(lat1)
	radLat2 := rad(lat2)
	a := radLat1 - radLat2
	b := rad(lon1) - rad(lon2)
	s := 2 * math.Asin(math.Sqrt(math.Pow(math.Sin(a / 2), 2) + math.Cos(radLat1) * math.Cos(radLat2) * math.Pow(math.Sin(b / 2), 2)))
	s = s * 6378137.0
	//s = Math.round(s * 10000) / 10000;
	return s
}

// 给定一个点，寻找离它最近的路口点
func FindIntersection(lat float64, lon float64) *defclass.Point {
	session := GetSesson()
	defer session.Close()
	db := session.DB(MAP_DB)
	dbcoll := db.C("mapIntersection")
	// 寻找其100米范围内的所有路口
	iter := dbcoll.Find(bson.M{"gis": bson.M{"$geoWithin": bson.M{"$centerSphere": []interface{}{[]float64{lon, lat}, ITHRESHOLD / 6378.137}}}}).Iter()
	min := math.MaxFloat64
	var res_id int64
	var res_lat, res_lon float64
	p := defclass.NewPoint0()
	// 从中找出最近的
	for iter.Next(p) {
		m := p.Gis.Map()
		latt := m["lat"].(float64)
		lonn := m["lon"].(float64)
		dis := Distance(lat, lon, latt, lonn)
		if dis < min {
			min = dis
			res_id = p.Id
			res_lat = latt
			res_lon = lonn
		}
	}
	// 如果没有与它相邻的路口
	if min == math.MaxFloat64 {
		return nil
	}
	point := defclass.NewPoint3(res_lat, res_lon, res_id)
	return point
}

// 将日期转换为秒
func ToSeconds(str string) int64 {
	tm, _ := time.ParseInLocation("2006-01-02 15:04:05.0", str, time.Local)
	return tm.Unix()
}

// 获取时
func GetHour(str string) int {
	tm, _ := time.ParseInLocation("2006-01-02 15:04:05.0", str, time.Local)
	return tm.Hour()
}

// 寻找所有的Trip
func findAllTrip(weeki string) []list.List {
	allTrip := make([]list.List, TIME_SLICE)
	for i := 0; i < TIME_SLICE; i++ {
		allTrip[i] = list.NewArrayList()
	}
	session := GetSesson()
	defer session.Close()
	db := session.DB(TRIP_DB)
	dbcoll := db.C(weeki)
	iter := dbcoll.Find(nil).Iter()
	tripData := defclass.NewTripData()
	// 遍历每一条出租车记录
	for iter.Next(tripData) {
		pickup_lat := tripData.Pickup["lat"].(float64)
		pickup_lon := tripData.Pickup["lon"].(float64)
		dropoff_lat := tripData.Dropoff["lat"].(float64)
		dropoff_lon := tripData.Dropoff["lon"].(float64)
		o := FindIntersection(pickup_lat, pickup_lon)
		if o == nil { // 如果没找到
			continue
		}
		d := FindIntersection(dropoff_lat, dropoff_lon)
		if d == nil {
			continue
		}
		if o.Id == d.Id { // 如果是回路,去除回路
			continue
		}
		time := ToSeconds(tripData.Dorpofftime) - ToSeconds(tripData.Pickuptime)
		trip := defclass.NewTrip4(o, d, tripData.Pickuptime, float64(time))
		trip.Distance = tripData.Distance // 实际行走长度
		hour := GetHour(tripData.Pickuptime)
		allTrip[hour].Add(trip)
	}
	return allTrip
}

// 合并等价Trip,同时去除太短或太长的路径
func MergeEquivalentTrip(weeki string) {
	allTrip := findAllTrip(weeki)
	session := GetSesson()
	defer session.Close()
	db := session.DB(TRIP_DB)
	for i := 0; i < TIME_SLICE; i++ {
		eqTrip := make(map[string]list.List, 0) // 用于存放等价Trip
		tripi := allTrip[i].Elements()
		for _, v := range tripi {
			trip := v.(*defclass.Trip)
			key := strconv.FormatInt(trip.GetO().Id, 10) + "_" + strconv.FormatInt(trip.GetD().Id, 10)
			if _, ok := eqTrip[key]; ok {
				eqTrip[key].Add(trip)
			} else {
				myList := list.NewArrayList()
				myList.Add(trip)
				eqTrip[key] = myList
			}
		}
		collName := "trip" + strconv.Itoa(i)
		dbcoll := db.C(collName)
		// 遍历eqTrip中的每一个等价类，进行合并
		for _, v := range eqTrip {
			o := v.Get(0).(*defclass.Trip).GetO()
			d := v.Get(0).(*defclass.Trip).GetD()
			n := v.Len()
			var travelTime float64 = 0.0
			var distance float64 = 0.0
			for j := 0; j < n; j++ {
				trip := v.Get(j).(*defclass.Trip)
				travelTime += trip.GetTravelTime()
				distance += trip.Distance
			}
			travelTime /= float64(n)	// 时间取平均值
			distance /= float64(n)	// 行驶距离取平均值
			//去除太短或太长的Trip,t<2min或t>1h
			if travelTime < 120 || travelTime > 3600 {
				continue
			}
			mytrip := defclass.NewSelectTrip(o.Id, d.Id, travelTime, distance)
			dbcoll.Insert(mytrip)
		}
	}
}