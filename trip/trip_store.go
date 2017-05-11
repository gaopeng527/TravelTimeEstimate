package trip

import (
	"TravelTimeEstimate/estimate"
	"strconv"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"TravelTimeEstimate/defclass"
	"os"
	"encoding/csv"
	"io"
	"math"
	"time"
)
// 删除选出的trip0-23，或者day1-31
func DropTrip(trip string, start int, end int) {
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.TRIP_DB)
	for i := start; i <= end; i++ {
		tripi := trip + strconv.Itoa(i)
		coll := db.C(tripi)
		coll.DropCollection()
		fmt.Println("删除集合" + tripi)
	}
}

// 分周存储
func StoreWeek(week int) {	// 0代表周日
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.TRIP_DB)
	dbcoll := db.C("trip-5")
	coll := db.C("week" + strconv.Itoa(week))
	re1 := "(\\d)"	// Any Single Digit 1
	re2 := "(\\d)"	// Any Single Digit 2
	re3 := "(\\d)"	// Any Single Digit 3
	re4 := "(\\d)"	// Any Single Digit 4
	re5 := "(-)"	// Any Single Character 1
	re6 := "(\\d)"	// Any Single Digit 5
	re7 := "(\\d)"	// Any Single Digit 6
	re8 := "(-)"	// Any Single Character 2
	for i := 1; i < 32; i++ {
		d := (i + 4) % 7
		if d != week {
			continue
		}
		prefix := re1+re2+re3+re4+re5+re6+re7+re8
		if i < 10 {
			prefix += "0" + strconv.Itoa(i)
		} else {
			prefix += strconv.Itoa(i)
		}
		dbcursor := dbcoll.Find(bson.M{"pickuptime": bson.M{"$regex": "^"+prefix+" .*$"}})
		iter := dbcursor.Iter()
		tripData := defclass.NewTripData()
		for iter.Next(tripData) {
			coll.Insert(tripData)
		}
		fmt.Println(i)
	}
}

// 存储某一天或几天的数据
func StoreDay(start int, end int) {
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.TRIP_DB)
	dbcoll := db.C("trip")
	re1 := "(\\d)"	// Any Single Digit 1
	re2 := "(\\d)"	// Any Single Digit 2
	re3 := "(\\d)"	// Any Single Digit 3
	re4 := "(\\d)"	// Any Single Digit 4
	re5 := "(-)"	// Any Single Character 1
	re6 := "(\\d)"	// Any Single Digit 5
	re7 := "(\\d)"	// Any Single Digit 6
	re8 := "(-)"	// Any Single Character 2
	for i := start; i <= end; i++ {
		coll := db.C("day" + strconv.Itoa(i))
		prefix := re1+re2+re3+re4+re5+re6+re7+re8
		if i < 10 {
			prefix += "0" + strconv.Itoa(i)
		} else {
			prefix += strconv.Itoa(i)
		}
		dbcursor := dbcoll.Find(bson.M{"pickuptime": bson.M{"$regex": "^"+prefix+" .*$"}})
		iter := dbcursor.Iter()
		tripData := defclass.NewTripData()
		for iter.Next(tripData) {
			coll.Insert(tripData)
		}
		fmt.Println(i)
	}
}

func MdbTrip(filePath string, collName string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.TRIP_DB)
	coll := db.C(collName)
	//----------------计时器--------------------//
	startMili := time.Now().Unix()	// 开始时间
	fmt.Println("开始时间：", startMili)
	//----------------计时器--------------------//
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return
		}
		//vendorID, _ := strconv.Atoi(record[0])
		tpep_pickup_datetime := record[1]+".0"
		tpep_dropoff_datetime := record[2]+".0"
		//passenger_count, _ := strconv.Atoi(record[3])
		trip_distance, _ := strconv.ParseFloat(record[4], 64)
		pickup_longitude, _ := strconv.ParseFloat(record[5], 64)
		pickup_latitude, _ := strconv.ParseFloat(record[6], 64)
		//ratecodeID, _ := strconv.Atoi(record[7])
		//store_and_fwd_flag := record[8]
		dropoff_longitude, _ := strconv.ParseFloat(record[9], 64)
		dropoff_latitude, _ := strconv.ParseFloat(record[10], 64)
		//payment_type, _ := strconv.Atoi(record[11])
		//fare_amount, _ :=  strconv.ParseFloat(record[12], 64)
		//extra, _ := strconv.ParseFloat(record[13], 64)
		//mta_tax, _ := strconv.ParseFloat(record[14], 64)
		//tip_amount, _ := strconv.ParseFloat(record[15], 64)
		//tolls_amount, _ := strconv.ParseFloat(record[16], 64)
		//improvement_surcharge, _ := strconv.ParseFloat(record[17], 64)
		//total_amount, _ := strconv.ParseFloat(record[18], 64)

		minlon := math.Min(pickup_longitude, dropoff_longitude)
		maxlon := math.Max(pickup_longitude, dropoff_longitude)
		minlat := math.Min(pickup_latitude, dropoff_latitude)
		maxlat := math.Max(pickup_latitude, dropoff_latitude)
		//去掉异常点然后存入Mongodb
		if(minlon>=MIN_LON&&maxlon<=MAX_LON&&minlat>=MIN_LAT&&maxlat<=MAX_LAT&&(pickup_latitude!=0||pickup_latitude!=0)&&(dropoff_longitude!=0||dropoff_latitude!=0)){
			tripData := defclass.NewTripData()
			tripData.Distance = trip_distance
			tripData.Dorpofftime = tpep_dropoff_datetime
			tripData.Dropoff = bson.M{"lat": dropoff_latitude, "lon": dropoff_longitude}
			tripData.Pickup = bson.M{"lat": pickup_latitude, "lon": pickup_longitude}
			tripData.Pickuptime = tpep_pickup_datetime
			coll.Insert(tripData)

		}
	}
	fmt.Println("数据插入完毕！")
	//----------------计时器--------------------//
	stopMili := time.Now().Unix()	// 结束时间
	fmt.Println("结束时间：", stopMili)
	fmt.Println("用时：", strconv.FormatInt((stopMili-startMili), 10)+"s")
	//----------------计时器--------------------//
}