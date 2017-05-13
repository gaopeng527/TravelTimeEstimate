package preprocess

import (
	"TravelTimeEstimate/estimate"
	"TravelTimeEstimate/defclass"
	"gopkg.in/mgo.v2"
	"fmt"
)
// 缩小地图点和打车请求的范围
type SmallPointAndTrip struct {
	min_lat float64
	max_lat float64
	min_lon float64
	max_lon float64
}

func NewSmallPointAndTrip(min_lat float64, max_lat float64, min_lon float64, max_lon float64) *SmallPointAndTrip {
	return &SmallPointAndTrip{min_lat: min_lat, max_lat: max_lat, min_lon: min_lon, max_lon: max_lon}
}

func (spat *SmallPointAndTrip) isValid(lat float64, lon float64) bool {
	if lat >= spat.min_lat && lat <= spat.max_lat && lon >= spat.min_lon && lon <= spat.max_lon {
		return true
	}
	return false
}

// 缩小地图点范围并存储到集合mapPointSmall
func (spat *SmallPointAndTrip) StorePoint() {
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.MAP_DB)
	coll := db.C("mapPoint")
	dbcoll := db.C("mapPointSmall")
	iter := coll.Find(nil).Iter()
	i := 0
	point := defclass.NewPoint0()
	for iter.Next(point) {
		m := point.Gis.Map()
		lat := m["lat"].(float64)
		lon := m["lon"].(float64)
		if spat.isValid(lat, lon) {
			dbcoll.Insert(point)
			i++
		}
	}
	index := mgo.Index{
		Key: []string{"$2dsphere:gis"},
		Bits: 26,
	}
	dbcoll.EnsureIndex(index)
	fmt.Println(i)
}

// 缩小Trip范围，并存储到tripSmall
func (spat *SmallPointAndTrip) StoreTrip() {
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.TRIP_DB)
	coll := db.C("trip")
	dbcoll := db.C("tripSmall")
	iter := coll.Find(nil).Iter()
	i := 0
	tripData := defclass.NewTripData()
	for iter.Next(tripData) {
		pickup_lat := tripData.Pickup["lat"].(float64)
		pickup_lon := tripData.Pickup["lon"].(float64)
		dropoff_lat := tripData.Dropoff["lat"].(float64)
		dropoff_lon := tripData.Dropoff["lon"].(float64)
		if spat.isValid(pickup_lat, pickup_lon) && spat.isValid(dropoff_lat, dropoff_lon) {
			dbcoll.Insert(tripData)
			i++
		}
	}
	fmt.Println(i)
}
