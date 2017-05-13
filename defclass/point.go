package defclass

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type Point struct{
	Id int64	 `bson:"_id"`
	lat float64
	lon float64
	Line_set []int64 `bson:"edge"` // 表示该点所属的弧集合，可能有多条
	Gis bson.D	 `bson:"gis"` // 经纬度信息
}

func NewPoint0() *Point {
	return &Point{}
}

func NewPoint1(p Point) *Point {
	return &Point{Id: p.Id, lat: p.lat, lon: p.lon}
}

func NewPoint2(lat float64, lon float64) *Point {
	return &Point{lat: lat, lon: lon}
}

func NewPoint3(lat float64, lon float64, id int64) *Point {
	return &Point{Id: id, lat: lat, lon: lon}
}

func (p *Point) GetLat() float64 {
	return p.lat
}

func (p *Point) SetLat(lat float64) {
	p.lat = lat
}

func (p *Point) GetLon() float64 {
	return p.lon
}

func (p *Point) SetLon(lon float64) {
	p.lon = lon
}

func (p *Point) Print() {
	fmt.Println(p.Id, p.lat, p.lon)
}