package driver

import (
	"io"
	"net"
)

type (
	Handler interface {
		Handle(net.Conn)
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
