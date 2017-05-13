package preprocess

import (
	"list"
	"TravelTimeEstimate/defclass"
	"TravelTimeEstimate/estimate"
	"set"
	Queue "container/list"
	"gopkg.in/mgo.v2"
	"fmt"
)
// 使图连通化，取最大连通子图
type Connected struct {
	box map[int64]list.List //存放邻接矩阵
	lineSet map[int64]*defclass.Line // 弧集
	pointSet map[int64]*defclass.Point // 点集
}

func NewConnected() *Connected {
	findStreet := estimate.NewFindStreet("mapPointSmall")
	lineSet := findStreet.GetStreet()
	pointSet := findStreet.GetPointSet()
	connected := &Connected{box: make(map[int64]list.List, 0), lineSet: lineSet, pointSet: pointSet}
	connected.buildMap()
	return connected
}

// 构建地图
func (connected *Connected) buildMap() {
	// 初始化邻接矩阵，无向图
	for _, line := range connected.lineSet {
		p := line.GetP()
		if _, ok := connected.box[p[0].Id]; ok {
			if !connected.box[p[0].Id].Contains(p[1].Id) {
				connected.box[p[0].Id].Add(p[1].Id)
			}
		} else {
			list := list.NewArrayList()
			list.Add(p[1].Id)
			connected.box[p[0].Id] = list
		}
		if _, ok := connected.box[p[1].Id]; ok {
			if !connected.box[p[1].Id].Contains(p[0].Id) {
				connected.box[p[1].Id].Add(p[0].Id)
			}
		} else {
			list := list.NewArrayList()
			list.Add(p[0].Id)
			connected.box[p[1].Id] = list
		}
	}
}

// 广度优先搜索
func (connected *Connected) bfs(root int64) set.Set {
	set := set.NewHashSet()
	queue := Queue.New()
	queue.PushBack(root)
	for queue.Len() > 0 {
		id := queue.Remove(queue.Front()).(int64)
		if set.Contains(id) {
			continue
		}
		set.Add(id)
		if _, ok := connected.box[id]; ok {
			n := connected.box[id].Len()
			for i := 0; i < n; i++ {
				pid := connected.box[id].Get(i).(int64)
				if !set.Contains(pid) {
					queue.PushBack(pid)
				}
			}
		}
	}
	return set
}

// 获取最大连通子图的点集
func (connected *Connected) getPoints() set.Set {
	var res set.Set = set.NewHashSet()
	myList := list.NewArrayList()
	for k, _ := range connected.pointSet {
		myList.Add(k)
	}
	// 采用广度优先搜索来查找
	for myList.Len() > res.Len() {
		mySet := connected.bfs(myList.Get(0).(int64))
		if mySet.Len() > res.Len() {
			res = mySet
		}
		a := mySet.Elements()
		for _, v := range a {
			id := v.(int64)
			myList.RemoveElement(id)
		}
	}
	return res
}

// 将最大连通子图存入mongodb
func (connected *Connected) StoreMaxConnect(fromColl string, toColl string) {
	myset := connected.getPoints()
	session := estimate.GetSesson()
	defer session.Close()
	db := session.DB(estimate.MAP_DB)
	coll := db.C(toColl)
	dbcoll := db.C(fromColl)
	a := myset.Elements()
	for _, v := range a {
		id := v.(int64)
		cursor := dbcoll.FindId(id)
		point := defclass.NewPoint0()
		cursor.One(point)
		coll.Insert(point)
	}
	index := mgo.Index{
		Key: []string{"$2dsphere:gis"},
		Bits: 26,
	}
	coll.EnsureIndex(index)
	fmt.Println("插入完毕：", myset.Len())
}
