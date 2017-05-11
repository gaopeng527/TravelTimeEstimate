package estimate

import "TravelTimeEstimate/defclass"

type MapLoc struct {
	PointSet map[int64]*defclass.Point
	LineSet map[int64]*defclass.Line
	PointNum int
	LineNum int
}

func NewMapLoc0() *MapLoc {
	return &MapLoc{}
}

func NewMapLoc2_1(pnum int, lnum int) *MapLoc {
	return &MapLoc{PointNum: pnum, PointSet: make(map[int64]*defclass.Point, 0), LineNum: lnum, LineSet: make(map[int64]*defclass.Line, 0)}
}

func NewMapLoc2_2(pointSet map[int64]*defclass.Point, lineSet map[int64]*defclass.Line) *MapLoc {
	return &MapLoc{PointSet: pointSet, LineSet: lineSet, PointNum: len(pointSet), LineNum: len(lineSet)}
}

func (mapLoc *MapLoc) AddPoint(id int64, lat float64, lon float64){
	mapLoc.PointSet[id] = defclass.NewPoint3(lat, lon, id)
}

func (mapLoc *MapLoc) AddLine(id int64, xid int64, yid int64, len float64, stid int64) {
	mapLoc.LineSet[id] = defclass.NewLine5(mapLoc.PointSet[xid], mapLoc.PointSet[yid], id, stid, len)
}