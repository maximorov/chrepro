package router

import (
	"bytes"
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

type TestRow struct {
	Id int32 `db:"id" json:"id"`
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
	chRep, err := net.Dial("tcp", "127.0.0.1:9001")
	handleError(err)

	//buf := make(chan []byte, 1024) // big buffer
	//tmp := make([]byte, 500)       // using small tmo buffer for demonstrating
	//
	//go func() {
	//	for {
	//		b := <-buf
	//		fmt.Printf("%x", b)
	//		conn.Write(b)
	//	}
	//}()
	//
	//for {
	//	n, err := conn.Read(tmp)
	//	if err != nil {
	//		if err != io.EOF {
	//			fmt.Println("read error:", err)
	//		}
	//		break
	//	}
	//	//fmt.Println("got", n, "bytes.")
	//	buf <- tmp[:n]
	//
	//}
	//fmt.Println("total size:", len(buf))
	//
	////for {
	//{
	//	go func() {
	//		buf := &bytes.Buffer{}
	//		mw := io.MultiWriter(buf, chRep)
	//		copied, err := io.Copy(mw, conn)
	//		if err != nil {
	//			log.Printf("Conection error: %s", err.Error())
	//		}
	//
	//		log.Printf("Connection closed. Bytes copied: %d", copied)
	//	}()
	//}
	//go func() {
	//	copied, err := io.Copy(conn, chRep)
	//	if err != nil {
	//		log.Printf("Connection error: %s", err.Error())
	//	}
	//
	//	log.Printf("Connection closed. Bytes copied: %d", copied)
	//}()
	////}
	//return
	stream := cio.NewStream(conn)
	clientHs := &proto2.ClientHandshake{}

	decoder := binary2.NewDecoder(stream)
	encoder := binary2.NewEncoder(stream)

	routeStream := cio.NewStream(chRep)

	_ = binary2.NewDecoder(routeStream)
	_ = binary2.NewEncoder(routeStream)

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

			n, err := routeStream.Write(decoder.FlushBufBytes())
			fmt.Printf("Wirtten: %d\n", n)
			handleError(err)
			err = routeStream.Flush()
			handleError(err)
			//res := make([]byte, 1024)
			//n, err = routeStream.Read(res)
			//handleError(err)
		case proto2.ClientPing:
			if err := encoder.Byte(proto2.ServerPong); err != nil {
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
		case proto2.ServerProgress:
			log.Fatal("Some progress in")
		case proto2.ClientQuery:
			q := proto2.Query{}
			if err := q.Decode(decoder /*, c.revision*/); err != nil {
				handleError(err)
			}

			for _, t := range r.targets {
				if t.forward(q.TableName) {
					log.Printf("Forward to %s:%s\n", t.host, t.port)

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
