package defclass

import "gopkg.in/mgo.v2/bson"

type TripData struct {
	Pickuptime string	`bson:"pickuptime"`
	Dorpofftime string	`bson:"dorpofftime"`
	Pickup bson.M		`bson:"pickup"`
	Dropoff bson.M		`bson:"dropoff"`
	Distance float64	`bson:"distance"`
}

func NewTripData() *TripData {
	return &TripData{}
}