package clickhouse

import (
	"fmt"
	"log"
	"net"
	"router/app/clickhouse/binary"
	"router/app/clickhouse/io"
	"router/app/clickhouse/proto"
	"router/app/driver"
	"time"
)

func NewHandler(rtp driver.RouteTargetsPool) *Handler {
	return &Handler{routes: rtp}
}

type Handler struct {
	routes  driver.RouteTargetsPool
	conn    net.Conn
	stream  *io.Stream
	closed  bool
	encoder *binary.Encoder
	decoder *binary.Decoder
}

// Handle natively copied new instance of Handler to process connection
func (s Handler) NewInstance(conn net.Conn) driver.Handler {
	s.conn = conn
	s.stream = io.NewStream(conn)
	s.decoder = binary.NewDecoder(s.stream)
	s.encoder = binary.NewEncoder(s.stream)

	return &s
}

// Handle natively copied new instance of Handler to process connection
func (s *Handler) Stop() error {
	if s.closed {
		return nil
	}
	s.closed = true
	s.encoder = nil
	s.decoder = nil
	if err := s.stream.Close(); err != nil {
		return err
	}
	if err := s.conn.Close(); err != nil {
		return err
	}
	return nil
}

// Handle natively copied new instance of Handler to process connection
func (s *Handler) Handle() {
	var helloMessage []byte
	for {
		packet, err := s.decoder.ReadByte()
		handleError(err)
		switch packet {
		case proto.ClientHello:
			clientHs := &proto.ClientHandshake{}
			if err := clientHs.Decode(s.decoder); err != nil {
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
			s.encoder.Byte(proto.ServerHello)

			err = serverHs.Encode(s.encoder)
			handleError(err)
			err = s.encoder.Flush()
			handleError(err)

			helloMessage = s.decoder.FlushBufBytes()

			//res := make([]byte, 1024)
			//n, err = routeStream.Read(res)
			//handleError(err)
		case proto.ClientPing:
			if err := s.encoder.Byte(proto.ServerPong); err != nil {
				handleError(err)
			}
			if err := s.encoder.Flush(); err != nil {
				handleError(err)
			}
		case proto.ServerProgress:
			log.Fatal("Some progress in")
		case proto.ClientQuery:
			q := proto.Query{}
			if err := q.Decode(s.decoder /*, c.revision*/); err != nil {
				handleError(err)
			}

			routeStream, _ := s.routes.Choose(q.Body)

			_, err := routeStream.Write(helloMessage)
			//fmt.Printf("Wirtten: %d\n", n)
			handleError(err)
			err = routeStream.Flush()
			handleError(err)

			go func() {
				n, err := routeStream.Write(s.decoder.FlushBufBytes())
				if n == 0 {
					time.Sleep(10 * time.Second)
				}
				//fmt.Printf("Wirtten: %d\n", n)
				handleError(err)
				err = routeStream.Flush() // send SQL request
				handleError(err)
			}()
			{
				{
					tmp := make([]byte, 36)
					_, _ = routeStream.Read(tmp)
					handleError(err)
				}
				{
					tmp := make([]byte, 31)
					_, err := routeStream.Read(tmp)
					//fmt.Printf("%x %x\n", tmp[29], tmp[30])
					_, _ = s.stream.Write(tmp)
					err = s.stream.Flush()
					handleError(err)
				}
				{
					tmp := make([]byte, 36)
					_, err := routeStream.Read(tmp)
					//fmt.Printf("%x %x\n", tmp[34], tmp[35])
					_, _ = s.stream.Write(tmp)
					err = s.stream.Flush()
					handleError(err)
				}
				for {
					tmp := make([]byte, 636) // TODO: fix it
					_, err := routeStream.Read(tmp)
					//for i := range tmp {
					//fmt.Printf("%x\n", tmp[i])
					//}
					handleError(err)
				}
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
