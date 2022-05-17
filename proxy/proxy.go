package proxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"router/clickhouse/lib/binary"
	"router/clickhouse/lib/io"
	"router/clickhouse/lib/proto"
	"time"
)

func NewProxy(host, port string, ctx context.Context) *Proxy {
	return &Proxy{
		host: host,
		port: port,
		ctx:  ctx,
	}
}

type Proxy struct {
	host           string
	port           string
	connectionId   uint64
	enableDecoding bool
	ctx            context.Context
	shutDownAsked  bool
}

func (r *Proxy) Start(port string) error {
	log.Printf("Start listening on: %s", port)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
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

func (r *Proxy) handle(conn net.Conn /*, connectionId uint64, enableDecoding bool*/) {
	stream := io.NewStream(conn)
	clientHs := &proto.ClientHandshake{}

	decoder := binary.NewDecoder(stream)
	encoder := binary.NewEncoder(stream)

	for {
		packet, err := decoder.ReadByte()
		handleError(err)
		switch packet {
		case proto.ClientHello:
			if err := clientHs.Decode(decoder); err != nil {
				handleError(err)
			}
			fmt.Printf("[handshake] <- %s\n", clientHs)

			serverHs := &proto.ServerHandshake{
				Name:        "TCP Router",
				DisplayName: "TCP Router",
				Revision:    proto.DBMS_TCP_PROTOCOL_VERSION,
				Timezone:    time.UTC,
			}
			serverHs.Version.Major = 1
			serverHs.Version.Minor = 1
			serverHs.Version.Patch = 0
			encoder.Byte(proto.ServerHello)

			err = serverHs.Encode(encoder)
			handleError(err)
			err = encoder.Flush()
			handleError(err)
		case proto.ClientPing:
			if err := encoder.Byte(proto.ServerPong); err != nil {
				handleError(err)
			}
			if err := encoder.Flush(); err != nil {
				handleError(err)
			}
		case proto.ServerProgress:
			log.Fatal("Some progress in")
		default:
			fmt.Errorf("[handshake] unexpected packet [%d] from server", packet)
		}

		log.Print("Handling completed\n")
	}
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
