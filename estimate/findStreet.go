package estimate

import (
	"TravelTimeEstimate/defclass"
)

type FindStreet struct {
	lineSet map[int64]*defclass.Line
	pointSet map[int64]*defclass.Point
	arcCollName string
}

func NewFindStreet(arcCollName string) *FindStreet {
	return &FindStreet{arcCollName: arcCollName}
}

// 从集合中获取地图线段信息
func (findStreet *FindStreet) GetStreet() map[int64]*defclass.Line {
	findStreet.lineSet = make(map[int64]*defclass.Line, 0)
	findStreet.pointSet = make(map[int64]*defclass.Point, 0)
	session := GetSesson()
	defer session.Close()
	db := session.DB(MAP_DB)
	dbcoll := db.C(findStreet.arcCollName)
	iter := dbcoll.Find(nil).Iter()
	p := defclass.NewPoint0()
	for iter.Next(p) {
		m := p.Gis.Map()
		lat := m["lat"].(float64)
		lon := m["lon"].(float64)
		point := defclass.NewPoint3(lat, lon, p.Id)
		findStreet.pointSet[p.Id] = point
		coll := db.C("mapArc")
		for _, sid := range p.Line_set {
			if _, ok := findStreet.lineSet[sid]; ok { // 如果该道路已经添加过
				continue
			}
			cursor := coll.FindId(sid)
			if cnt, _ := cursor.Count(); cnt > 0 {
				arc := defclass.NewLine0()
				cursor.One(arc)
				x := arc.Gis["x"].(int64)
				y := arc.Gis["y"].(int64)
				p1 := defclass.NewPoint0()
				p2 := defclass.NewPoint0()
				//查找第一个顶点
				pcursor := dbcoll.FindId(x)
				if cnt, _ := pcursor.Count(); cnt == 0 { //如果没找到
					continue
				}
				p1.Id = x
				tempP := defclass.NewPoint0()
				pcursor.One(tempP)
				tempM := tempP.Gis.Map()
				p1.SetLat(tempM["lat"].(float64))
				p1.SetLon(tempM["lon"].(float64))
				//查找第二个顶点
				pcursor = dbcoll.FindId(y)
				if cnt, _ := pcursor.Count(); cnt == 0 { //如果没找到
					continue
				}
				p2.Id = y
				pcursor.One(tempP)
				tempM = tempP.Gis.Map()
				p2.SetLat(tempM["lat"].(float64))
				p2.SetLon(tempM["lon"].(float64))
				line := defclass.NewLine5(p1, p2, sid, arc.Wayid, arc.Length)
				// 初始化道路行驶时间
				line.SetTravelTime(line.Length/VINIT)
				findStreet.lineSet[sid] = line
			}
		}
	}
	return findStreet.lineSet
}

// 获取地图点集
func (findStreet *FindStreet) GetPointSet() map[int64]*defclass.Point {
	if findStreet.pointSet == nil {
		findStreet.GetStreet()
	}
	return findStreet.pointSet
}