package defclass

type Trip struct {
	o *Point //起始路口
	d *Point //结束路口
	Sid int64 //起始路口id
	Eid int64 //终点路口id
	Et float64 //估计的行车时间
	startTime string //起始时间
	travelTime float64 //行车时间
	Length float64 // 总长度
	Distance float64 // 实际行走长度
	path []interface{} // 最短路径
}

func NewTrip0() *Trip {
	return &Trip{}
}

func (trip *Trip) GetPath() []interface{} {
	return trip.path
}

func (trip *Trip) SetPath(path []interface{}) {
	trip.path = path
}

func NewTrip4(o *Point, d *Point, startTime string, travelTime float64) *Trip {
	return &Trip{o: o, d: d, startTime: startTime, travelTime: travelTime}
}

func NewTrip3(o *Point, d *Point, travelTime float64) *Trip {
	return &Trip{o: o, d: d, travelTime: travelTime}
}

func (trip *Trip) GetO() *Point {
	return trip.o
}

func (trip *Trip) SetO(o *Point) {
	trip.o = o
}

func (trip *Trip) GetD() *Point {
	return trip.d
}

func (trip *Trip) SetD(d *Point) {
	trip.d = d
}

func (trip *Trip) GetStartTime() string {
	return trip.startTime
}

func (trip *Trip) SetStartTime(startTime string) {
	trip.startTime = startTime
}

func (trip *Trip) GetTravelTime() float64 {
	return trip.travelTime
}

func (trip *Trip) SetTravelTime(travelTime float64) {
	trip.travelTime = travelTime
}