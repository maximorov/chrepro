package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"router/app/driver"
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

	go func() {
		log.Printf("Waiting for shut down signal ^C")
		<-r.ctx.Done()
		r.shutDownAsked = true
		log.Printf("Shut down signal received, closing connections...")
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		handleError(err)
		r.connectionId += 1
		if err != nil {
			log.Printf("Failed to accept new connection: [%d] %s", r.connectionId, err.Error())
			if r.shutDownAsked {
				log.Printf("Shutdown asked [%d]", r.connectionId)
				break
			}
			continue
		}

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
	h := r.handler.NewInstance(conn)
	go func(h driver.Handler) {
		<-r.ctx.Done()
		log.Printf("Shut down signal received, Stoping handler...")
		handleError(h.Stop())
	}(h)
	h.Handle()

}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
