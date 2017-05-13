package defclass

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

type Line struct {
	p []*Point
	Length float64	`bson:"length"`
	Sid int64	`bson:"_id"`
	Wayid int64	`bson:"wayid"`
	travelTime float64
	Gis bson.M	 `bson:"gis"` // 线的两端点id，"gis" : { "y" : 84047856 ,"x" : 716309074}
}

func NewLine0() *Line {
	return &Line{p: make([]*Point, 2)}
}

func NewLine4(p1 *Point, p2 *Point, sid int64, length float64) *Line {
	p := make([]*Point, 2)
	p[0] = p1
	p[1] = p2
	return &Line{p: p, Sid: sid, Length: length}
}

func NewLine5(p1 *Point, p2 *Point, sid int64, wayid int64, length float64) *Line {
	p := make([]*Point, 2)
	p[0] = p1
	p[1] = p2
	return &Line{p: p, Sid: sid, Length: length, Wayid: wayid}
}

func (line *Line) GetP() []*Point {
	return line.p
}

func (line *Line) SetP(p []*Point) {
	line.p = p
}

func (line *Line) GetTravelTime() float64 {
	return line.travelTime
}

func (line *Line) SetTravelTime(travelTime float64) {
	line.travelTime = travelTime
}

func (line *Line) Print() {
	fmt.Println(line.Sid, line.p[0].Id, line.p[1].Id, line.Length, line.travelTime)
}

