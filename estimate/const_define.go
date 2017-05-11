package estimate

// dijkstra
const COST = 0.0 // 最短路径计算时的换路代价
const INF = 2100000000.0; //设置无穷取值
// findStreet
const VINIT = 15.0; // 道路初始速度，单位m/s
const MAP_DB = "MapLocNYCTime" // 地图数据库名称
//const SHORTEST_ROUTE_DB = "Test" // 最短路径数据库名称
// findTrip
const TRIP_DB = "MapLocNYCTime" // 旅途数据库名称
const ITHRESHOLD  = 0.1 // 将上下车地点定位到多大范围内的路口,单位km
const PI  = 3.141592653589793238462643383279502884
const TIME_SLICE  = 24 //将一天分为的时段数
// algorithm
const USE_WEEK = "week0"
const ES_WEEK_DAY = "day28"
const ERROR = 0.06