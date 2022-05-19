package main

import (
	"context"
	"log"
	"router/app"
	"router/app/clickhouse"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	targetsPool := app.NewRouteTargetsPool(app.NewRouteTarget("127.0.0.1", "9001",
		func(ts []string) bool {
			for i := range ts {
				if ts[i] == `t1` {
					return true
				}
			}
			return false
		}), app.NewRouteTarget("127.0.0.1", "9002",
		func(ts []string) bool {
			for i := range ts {
				if ts[i] == `t2` {
					return true
				}
			}
			return false
		}))

	handler := clickhouse.NewHandler(targetsPool)
	cnf := app.NewConfig("127.0.0.1", "9000")
	srv := app.NewRouter(ctx, cnf, handler)

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
