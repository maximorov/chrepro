package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"router/app/driver"
	"time"
)

type Router struct {
	ctx            context.Context
	listenHost     string
	listenPort     string
	handler        driver.Handler
	connectionId   uint64
	enableDecoding bool
	shutDownAsked  bool
}

func NewRouter(ctx context.Context, c Config, h driver.Handler) *Router {
	return &Router{ctx: ctx,
		listenHost: c.listenHost,
		listenPort: c.listenPort,
		handler:    h,
	}
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

	go func() {
		time.Sleep(time.Second * 5)
		ln.Close()
	}()

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
	defer func() {
		err := conn.Close()
		handleError(err)
	}()
	r.handler.Handle(conn)
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

}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
