package clickhouse

import (
	"bytes"
	"fmt"
	"log"
	"net"
	binary2 "router/app/clickhouse/binary"
	"router/app/clickhouse/io"
	"router/app/clickhouse/proto"
	"router/app/driver"
	"time"
)

func NewHandler(rtp driver.RouteTargetsPool) *Handler {
	return &Handler{rtp}
}

type Handler struct {
	pool driver.RouteTargetsPool
}

func (s *Handler) Handle(conn net.Conn) {
	stream := io.NewStream(conn)
	clientHs := &proto.ClientHandshake{}

	decoder := binary2.NewDecoder(stream)
	encoder := binary2.NewEncoder(stream)

	var helloMessage []byte

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

			helloMessage = decoder.FlushBufBytes()

			//res := make([]byte, 1024)
			//n, err = routeStream.Read(res)
			//handleError(err)
		case proto.ClientPing:
			if err := encoder.Byte(proto.ServerPong); err != nil {
				handleError(err)
			}
			if err := encoder.Flush(); err != nil {
				handleError(err)
			}

			//n, err := routeStream.Write(decoder.FlushBufBytes())
			//fmt.Printf("Wirtten: %d\n", n)
			//handleError(err)
			//err = routeStream.Flush()
			//handleError(err)
			//res := make([]byte, 1024)
			//n, err = routeStream.Read(res)
			//handleError(err)
		case proto.ServerProgress:
			log.Fatal("Some progress in")
		case proto.ClientQuery:
			q := proto.Query{}
			if err := q.Decode(decoder /*, c.revision*/); err != nil {
				handleError(err)
			}

			routeStream, _ := s.pool.Choose(q.Body)

			n, err := routeStream.Write(helloMessage)
			fmt.Printf("Wirtten: %d\n", n)
			handleError(err)
			err = routeStream.Flush()
			handleError(err)

			go func() {
				n, err := routeStream.Write(decoder.FlushBufBytes())
				if n == 0 {
					time.Sleep(10 * time.Second)
				}
				fmt.Printf("Wirtten: %d\n", n)
				handleError(err)
				err = routeStream.Flush() // send SQL request
				handleError(err)
			}()
			{
				buf := &bytes.Buffer{}
				{
					tmp := make([]byte, 36)
					_, _ = routeStream.Read(tmp)
					handleError(err)
				}
				{
					tmp := make([]byte, 31)
					_, err := routeStream.Read(tmp)
					fmt.Printf("%x %x\n", tmp[29], tmp[30])
					_, _ = stream.Write(tmp)
					err = stream.Flush()
					handleError(err)
				}
				{
					tmp := make([]byte, 36)
					_, err := routeStream.Read(tmp)
					fmt.Printf("%x %x\n", tmp[34], tmp[35])
					_, _ = stream.Write(tmp)
					err = stream.Flush()
					handleError(err)
				}
				for {
					tmp := make([]byte, 31)
					_, err := routeStream.Read(tmp)
					fmt.Printf("%x\n", tmp[30])
					handleError(err)
				}
				_, _ = stream.Write(buf.Bytes())
				err = stream.Flush()
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
