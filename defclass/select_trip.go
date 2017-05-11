package defclass
type SelectTrip struct {
	Sid int64		`bson:"sid"`
	Eid int64		`bson:"eid"`
	TravelTime float64	`bson:"travelTime"`
	Distance float64	`bson:"distance"`
}

func NewSelectTrip(sid int64, eid int64, travelTime float64, distance float64) *SelectTrip {
	return &SelectTrip{Sid: sid, Eid: eid, TravelTime: travelTime, Distance: distance}
}

func NewSelectTrip0() *SelectTrip {
	return &SelectTrip{}
}