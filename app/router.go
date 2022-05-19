package app

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"router/app/lib/binary"
	"router/app/lib/io"
	"router/app/lib/proto"
	"time"
)

func NewRouter(ctx context.Context, c Config) *Router {
	return &Router{ctx: ctx, Config: c}
}

type TestRow struct {
	Id int32 `db:"id" json:"id"`
}

type Router struct {
	Config
	ctx            context.Context
	connectionId   uint64
	enableDecoding bool
	shutDownAsked  bool
}

func (r *Router) Start() error {
	log.Printf("Start listening on: %s:%s", r.listenHost, r.listenPort)
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", r.listenHost, r.listenPort))
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
	stream := io.NewStream(conn)
	clientHs := &proto.ClientHandshake{}

	decoder := binary.NewDecoder(stream)
	encoder := binary.NewEncoder(stream)

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

			routeStream, _ := r.targetsPool.Choose(q.Body)

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
