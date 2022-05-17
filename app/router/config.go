package router

type Config struct {
	listenHost string
	listenPort string
	targets    []RouteTarget
}

func NewConfig(h, p string, rt ...RouteTarget) Config {
	return Config{h, p, rt}
}

type RouteRule func(tName string) bool

type RouteTarget struct {
	host    string
	port    string
	forward RouteRule
}

func NewRouteTarget(h, p string, fn RouteRule) RouteTarget {
	return RouteTarget{h, p, fn}
}
