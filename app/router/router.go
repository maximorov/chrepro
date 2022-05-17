package router

import (
	"context"
	"fmt"
	"log"
	"net"
	binary2 "router/app/clickhouse/lib/binary"
	cio "router/app/clickhouse/lib/io"
	proto2 "router/app/clickhouse/lib/proto"
	"time"
)

func New(ctx context.Context, c Config) *Router {
	return &Router{
		host:    c.listenHost,
		port:    c.listenPort,
		targets: c.targets,
		ctx:     ctx,
	}
}

type Router struct {
	host           string
	port           string
	targets        []RouteTarget
	connectionId   uint64
	enableDecoding bool
	ctx            context.Context
	shutDownAsked  bool
}

func (r *Router) Start() error {
	log.Printf("Start listening on: %s:%s", r.host, r.port)
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", r.host, r.port))
	if err != nil {
		return err
	}

	//go func() {
	//	log.Printf("Waiting for shut down signal ^C")
	//	<-r.ctx.Done()
	//	r.shutDownAsked = true
	//	log.Printf("Shut down signal received, closing connections...")
	//	ln.Close()
	//}()

	for {
		conn, err := ln.Accept()
		handleError(err)
		//r.connectionId += 1
		//if err != nil {
		//	log.Printf("Failed to accept new connection: [%d] %s", r.connectionId, err.Error())
		//	if r.shutDownAsked {
		//		log.Printf("Shutdown asked [%d]", r.connectionId)
		//		break
		//	}
		//	continue
		//}

		log.Printf("Connection accepted: [%d] %s\n", r.connectionId, conn.RemoteAddr())
		go r.handle(conn /*, r.connectionId, r.enableDecoding*/)
	}

	return nil
}

func (r *Router) handle(conn net.Conn /*, connectionId uint64, enableDecoding bool*/) {
	stream := cio.NewStream(conn)
	clientHs := &proto2.ClientHandshake{}

	decoder := binary2.NewDecoder(stream)
	encoder := binary2.NewEncoder(stream)

	for {
		packet, err := decoder.ReadByte()
		handleError(err)
		switch packet {
		case proto2.ClientHello:
			if err := clientHs.Decode(decoder); err != nil {
				handleError(err)
			}
			fmt.Printf("[handshake] <- %s\n", clientHs)

			serverHs := &proto2.ServerHandshake{
				Name:        "TCP Router",
				DisplayName: "TCP Router",
				Revision:    proto2.DBMS_TCP_PROTOCOL_VERSION,
				Timezone:    time.UTC,
			}
			serverHs.Version.Major = 1
			serverHs.Version.Minor = 1
			serverHs.Version.Patch = 0
			encoder.Byte(proto2.ServerHello)

			err = serverHs.Encode(encoder)
			handleError(err)
			err = encoder.Flush()
			handleError(err)
		case proto2.ClientPing:
			if err := encoder.Byte(proto2.ServerPong); err != nil {
				handleError(err)
			}
			if err := encoder.Flush(); err != nil {
				handleError(err)
			}
		case proto2.ServerProgress:
			log.Fatal("Some progress in")
		case proto2.ClientQuery:
			q := proto2.Query{}
			if err := q.Decode(decoder /*, c.revision*/); err != nil {
				handleError(err)
			}
		default:
			fmt.Errorf("[handshake] unexpected packet [%d] from server", packet)
		}
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//buf := &bytes.Buffer{}
//copied, err := io.Copy(buf, conn)
//fmt.Printf("%v %v\n", copied, err)
////for _, t := range r.targets {
////	if t.forward(q.TableName) {
////		log.Printf("Forward to %s:%s\n", t.host, t.port)
////		cn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", t.host, t.port))
////		handleError(err)
////		n, err := cn.Write(buf.Bytes())
////		handleError(err)
////		fmt.Printf("Req: %v\n", n)
////		var ans []byte
////		n, err = cn.Read(ans)
////		handleError(err)
////		fmt.Printf("Ans: %v %v", n, ans)
////	}
////}
