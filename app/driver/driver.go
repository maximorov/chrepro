package driver

import (
	"io"
	"net"
)

type (
	Handler interface {
		NewInstance(conn net.Conn) Handler
		Handle()
		Stop() error
	}
	RouteTargetsPool interface {
		Choose(SQLQuery string) (Stream, error)
	}
	Stream interface {
		io.ReadWriteCloser
		Compress(v bool)
		Flush() error
	}
)
