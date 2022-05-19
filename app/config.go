package app

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"router/app/clickhouse/io"
	"router/app/driver"
)

type Config struct {
	listenHost string
	listenPort string
}

func NewConfig(h, p string) Config {
	return Config{h, p}
}

type RouteRule func(tName []string) bool

type RouteTarget struct {
	host    string
	port    string
	forward RouteRule
}

func NewRouteTarget(h, p string, fn RouteRule) RouteTarget {
	return RouteTarget{h, p, fn}
}

type RouteTargetsPool struct {
	targets []RouteTarget
}

func NewRouteTargetsPool(rts ...RouteTarget) *RouteTargetsPool {
	targets := make([]RouteTarget, 0, len(rts))
	targets = append(targets, rts...)

	return &RouteTargetsPool{targets}
}

func (s *RouteTargetsPool) Choose(SQLQuery string) (driver.Stream, error) {
	var chosenStream *io.Stream

	rxp, _ := regexp.Compile(`FROM ([_\-\w0-9]+)`)
	tname := rxp.FindSubmatch([]byte(SQLQuery))
	fmt.Printf("%v\n", string(tname[1]))
	tableName := string(tname[1])

	for _, t := range s.targets {
		if t.forward([]string{tableName}) {
			chRep, err := net.Dial("tcp", "127.0.0.1:9001")
			handleError(err)
			chosenStream = io.NewStream(chRep)
			log.Printf("Forward to %s:%s\n", t.host, t.port)
			break
		}
	}

	return chosenStream, nil
}
