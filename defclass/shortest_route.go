package defclass

import (
	"bytes"
	"fmt"
)

type ShortestRoute struct {
	Id    string       	    `bson:"_id"`
	Path  []interface{}        `bson:"path"` //bson:"name" 表示mongodb数据库中对应的字段名称
	Length float64              `bson:"length"`
}

func NewShortestRoute() *ShortestRoute{
	return &ShortestRoute{}
}

func (route *ShortestRoute) String() string{
	var buf bytes.Buffer
	buf.WriteString("ShortestRoute{")
	buf.WriteString(fmt.Sprintf("%v, ", route.Id))
	buf.WriteString(fmt.Sprintf("%v, ", route.Path))
	buf.WriteString(fmt.Sprintf("%v}", route.Length))
	return buf.String()
}
