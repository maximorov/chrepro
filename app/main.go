package main

import (
	"context"
	"log"
	"router/app/router"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	cnf := router.NewConfig("127.0.0.1", "9000",
		router.NewRouteTarget("127.0.0.1", "9001",
			func(t string) bool {
				return t == `t1`
			}),
		router.NewRouteTarget("127.0.0.1", "9002",
			func(t string) bool {
				return t == `t2`
			}))
	srv := router.New(ctx, cnf)

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt)
	//go func() {
	//	for sig := range c {
	//		log.Printf("Signal received %v, stopping and exiting...", sig)
	//		cancel()
	//	}
	//}()

	err := srv.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// Router
// MiniServer
//
