package estimate

import (
	"list"
	"time"
	"TravelTimeEstimate/defclass"
	"reflect"
	"fmt"
	"strconv"
)

type Dijkstra struct {
	box map[int64]list.List //存放邻接矩阵
	visO map[int64]int64 // 判断该点是否作为起点被访问过
	visD map[int64]int64 // 判断该点是否作为终点被访问过
	dis map[int64]float64 // 存放到初始点的距离
	cost map[int64]float64 // 存放换路的代价
	flagCost map[int64]int64 // 用于标识当前换路代价的版本
	mapLoc *MapLoc // 存放地图信息
}

func NewDijkstra(mapLoc *MapLoc) *Dijkstra {
	dij := &Dijkstra{box: make(map[int64]list.List, 0), visO: make(map[int64]int64, 0),
	visD: make(map[int64]int64, 0), dis: make(map[int64]float64, 0), cost: make(map[int64]float64, 0),
	flagCost: make(map[int64]int64, 0), mapLoc: mapLoc}
	dij.buildMap()
	return dij
}

// 边
type Edge struct {
	Pid int64
	Len float64
	Sid int64 // 边的id
}

func NewEdge(pid int64, len float64, sid int64) *Edge {
	return &Edge{Pid: pid, Len: len, Sid: sid}
}

// 节点
type Node struct {
	To int64
	Len float64
	PathId list.List
	PathvId list.List
}

func NewNode0() *Node {
	return &Node{}
}

func NewNode2(to int64, len float64) *Node {
	return &Node{To: to, Len: len}
}

func NewNode4(to int64, len float64, pathId list.List, pathvId list.List) *Node {
	return &Node{To: to, Len: len, PathId: pathId, PathvId: pathvId}
}

// 构建地图
func (dij *Dijkstra) buildMap() {
	// 初始化邻接矩阵，无向图
	for _, line := range dij.mapLoc.LineSet {
		p := line.GetP()
		if _, ok := dij.box[p[0].Id]; ok {
			if !dij.box[p[0].Id].Contains(p[1].Id){
				dij.box[p[0].Id].Add(NewEdge(p[1].Id, line.Length, line.Sid))
			}
		}else {
			myList := list.NewArrayList()
			myList.Add(NewEdge(p[1].Id, line.Length, line.Sid))
			dij.box[p[0].Id] = myList
		}

		if _, ok := dij.box[p[1].Id]; ok {
			if !dij.box[p[1].Id].Contains(p[0].Id) {
				dij.box[p[1].Id].Add(NewEdge(p[0].Id, line.Length, line.Sid))
			}
		} else {
			mylist := list.NewArrayList()
			mylist.Add(NewEdge(p[0].Id, line.Length, line.Sid))
			dij.box[p[1].Id] = mylist
		}
	}
	// 初始化dis，vis信息
	for pid := range dij.mapLoc.PointSet {
		dij.dis[pid] = INF
		dij.visO[pid] = 0
		dij.visD[pid] = 0
		dij.cost[pid] = 0.0 // 初始化换路代价为0
		dij.flagCost[pid] = 0
	}
}

// 求S到T的最短路 避免初始化，防止多线程时造成错误，进行同步，返回最短路经过的路段id
func (dij *Dijkstra) Solve(S int64, T int64) (float64, []interface{}) {
	// 存放起点到各点的最短路径id
	pathId := make(map[int64]list.List, 0)
	if S == T {
		return 0.0, make([]interface{}, 0)
	}
	startTime := time.Now().Unix()
	// 判断是否已经计算过
	session := GetSesson()
	defer session.Close()
	db := session.DB(MAP_DB)
	dbcoll := db.C("shortestRoute")
	cursor := dbcoll.FindId(strconv.FormatInt(S, 10) + "-" + strconv.FormatInt(T, 10))
	if cnt, _ := cursor.Count(); cnt > 0 {
		route := defclass.NewShortestRoute()
		cursor.One(route)
		length := route.Length
		return length, route.Path
	}
	pq := NewPriorityQueue(reflect.TypeOf(NewNode0()))
	dij.dis[S] = 0.0
	dij.cost[S] = 0.0
	timestamp := time.Now().UnixNano() // 避免重新初始化
	dij.flagCost[S] = timestamp
	pq.Push(NewItem(NewNode2(S, 0.0), 0.0))
	for pq.Len() > 0 {
		// 如果时间超过30s没有算出最短路就跳出循环
		if time.Now().Unix() - startTime > 30 {
			break
		}
		now := pq.Pop().value.(*Node)
		po := now.To
		if po == T {
			break
		}
		if dij.visO[po] == timestamp {
			continue
		}
		dij.visO[po] = timestamp
		if dij.box[po] == nil {
			continue
		}
		var pathpoId list.List
		if _, ok := pathId[po]; ok {
			pathpoId = pathId[po]
		} else {
			pathpoId = list.NewArrayList()
			pathId[po] = pathpoId
		}
		n := dij.box[po].Len()
		for i := 0; i < n; i++ {
			edge := dij.box[po].Get(i).(*Edge)
			v := edge.Pid // 该弧的终点
			len := edge.Len
			var newCost float64
			if dij.flagCost[po] != timestamp {
				dij.cost[po] = 0.0
				dij.flagCost[po] = timestamp
			}
			if pathpoId.Len() == 0 {
				newCost = dij.cost[po]
			} else {
				sid := pathpoId.Get(pathpoId.Len() - 1).(int64)
				wayid1 := dij.mapLoc.LineSet[sid].Wayid
				wayid2 := dij.mapLoc.LineSet[edge.Sid].Wayid
				if wayid1 == wayid2 {
					newCost = dij.cost[po]
				} else {
					newCost = dij.cost[po] + COST
				}
			}
			if dij.flagCost[v] != timestamp {
				dij.cost[v] = 0.0
				dij.flagCost[v] = timestamp
			}
			if dij.visD[v] != timestamp || len + now.Len + newCost < dij.dis[v] + dij.cost[v] {
				pathvId := list.NewArrayList()
				list.AddList(pathvId, pathpoId)
				pathvId.Add(edge.Sid)
				pathId[v] = pathvId
				dij.visD[v] = timestamp
				dij.dis[v] = len + now.Len
				dij.cost[v] = newCost
				pq.Push(NewItem(NewNode2(v, len + now.Len + newCost), len + now.Len + newCost))
			}
		}
	}
	if _, ok := pathId[T]; ok {
		// 将结果插入数据库
		route := defclass.NewShortestRoute()
		route.Id = strconv.FormatInt(S, 10) + "-" + strconv.FormatInt(T, 10)
		route.Path = pathId[T].Elements()
		route.Length = dij.dis[T]
		dbcoll.Insert(route)
		return dij.dis[T], route.Path
	}
	return INF, nil
}

func (dij *Dijkstra) PrintPathInfo(sid int64, eid int64) {
	session := GetSesson()
	defer session.Close()
	db := session.DB(MAP_DB)
	dbcoll := db.C("shortestRoute")
	route := defclass.NewShortestRoute()
	dbcoll.FindId(strconv.FormatInt(sid, 10) + "-" + strconv.FormatInt(eid, 10)).One(route)
	fmt.Println(route)
}