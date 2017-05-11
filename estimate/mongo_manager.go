package estimate

import "gopkg.in/mgo.v2"

var session *mgo.Session
func GetSesson() *mgo.Session {
	if session == nil {
		session, _ = mgo.Dial("127.0.0.1:27017")
		session.SetMode(mgo.Monotonic, true)
		session.SetCursorTimeout(0)
	}
	return session.Copy()
}

func CloseSession() {
	if session != nil {
		session.Close()
	}
}

