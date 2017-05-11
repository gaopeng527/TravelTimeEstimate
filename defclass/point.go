package defclass

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type Point struct{
	Id int64	 `bson:"_id"`
	Lat float64
	Lon float64
	Line_set []int64 `bson:"edge"` // 表示该点所属的弧集合，可能有多条
	Gis bson.M	 `bson:"gis"` // 经纬度信息
}

func NewPoint0() *Point {
	return &Point{}
}

func NewPoint1(p Point) *Point {
	return &Point{Id: p.Id, Lat: p.Lat, Lon: p.Lon}
}

func NewPoint2(lat float64, lon float64) *Point {
	return &Point{Lat: lat, Lon: lon}
}

func NewPoint3(lat float64, lon float64, id int64) *Point {
	return &Point{Id: id, Lat: lat, Lon: lon}
}

func (p *Point) Print() {
	fmt.Println(p.Id, p.Lat, p.Lon)
}